use std::{
    cmp::{max, min},
    str::FromStr as _,
};

use anyhow::anyhow;
use orchard::{builder::BundleType, bundle::Flags, keys::OutgoingViewingKey, value::NoteValue};
use rand_core::{OsRng, RngCore, SeedableRng};
use sapling_crypto::{note_encryption::Zip212Enforcement, Anchor};
use secp256k1::{PublicKey, SecretKey};
use zcash_keys::{
    address::{Address, Receiver},
    encoding::AddressCodec,
};
use zcash_primitives::{
    legacy::{Script, TransparentAddress},
    transaction::{
        builder::{BuildConfig, Builder as TxBuilder},
        components::{transparent::builder::TransparentBuilder, OutPoint, TxOut},
        fees::zip317,
        sighash::{signature_hash, SignableInput, SIGHASH_ALL},
        txid::TxIdDigester,
        TransactionData, TxVersion,
    },
};
use zcash_proofs::prover::LocalTxProver;
use zcash_protocol::{
    consensus::{BlockHeight, BranchId},
    memo::{Memo, MemoBytes},
    value::{ZatBalance as Amount, Zatoshis},
};

use crate::{
    addr::{get_ovk, get_vault_address, validate_address},
    config::Context,
    to_ba, to_hash, to_zcasherror, uniffi_async_export, uniffi_export,
    wallet::UTXO,
    ZcashError,
};

pub struct TxBytes {
    pub txid: String,
    pub data: Vec<u8>,
}

const BASE_FEE: u64 = 5_000;

pub fn send_to_vault(
    expiry_height: u32,
    sk: Vec<u8>,
    from: String,
    vault: Vec<u8>,
    amount: u64,
    memo: String,
) -> Result<TxBytes, ZcashError> {
    uniffi_async_export!(context, {
        // user inputs should be checked
        let sk =
            SecretKey::from_slice(&sk).map_err(to_zcasherror(anyhow!("Invalid Secret Key")))?;
        let valid_address = validate_address(from.clone())?;
        if !valid_address {
            return Err(ZcashError::InvalidAddress(from.clone()));
        }
        let to_addr = get_vault_address(vault)?;
        Zatoshis::from_u64(amount)
            .map_err(to_zcasherror(anyhow!("Invalid amount: {amount} zats")))?;
        if memo.len() > 80 {
            return Err(ZcashError::AssertError(
                anyhow!("Memo too long: {memo}").to_string(),
            ));
        }
        let utxos = crate::wallet::list_utxos_async(&context, from.clone()).await?;
        let (inputs, change, _) = select_utxos(&utxos, amount, &memo)?;

        let txb = pay_with_utxos(
            &context,
            expiry_height,
            sk,
            inputs,
            from,
            to_addr,
            amount,
            change,
            memo,
        );
        Ok::<_, ZcashError>(txb)
    })?
}

fn select_utxos(
    utxos: &[UTXO],
    amount: u64,
    memo: &str,
) -> Result<(Vec<UTXO>, u64, u64), ZcashError> {
    let num_touts: u64 = 2 + {
        if memo.is_empty() {
            0
        } else {
            let len = memo.len() + 2; // size in bytes of the OP_RETURN output
            ((len + 33) / 34) as u64 // 34 is the size of a PKH output
        }
    }; // vault + change
    let mut num_tins: u64 = 0;
    let fee = |num_tins, num_touts| max(num_tins, num_touts) * BASE_FEE;
    let mut current_fee = 0;

    let mut needed = amount;
    let mut inputs = vec![];
    let mut input_amount = 0;

    for utxo in utxos {
        let new_fee = fee(num_tins, num_touts);
        needed += new_fee - current_fee;
        current_fee = new_fee;

        let available = min(utxo.value, needed);
        needed -= available;
        input_amount += utxo.value;
        num_tins += 1;
        inputs.push(utxo.clone());

        if needed == 0 {
            break;
        }
    }
    if needed > 0 {
        return Err(ZcashError::NotEnoughFunds);
    }
    let f = fee(num_tins, num_touts);
    let change = input_amount - amount - f;

    Ok((inputs, change, f))
}

pub struct Output {
    pub address: String,
    pub amount: u64,
    pub memo: String,
}

pub struct PartialTx {
    pub height: u32,
    pub inputs: Vec<UTXO>,
    pub outputs: Vec<Output>,
    pub fee: u64,
    pub tx_seed: Vec<u8>,
}

pub fn pay_from_vault(
    height: u32,
    vault: Vec<u8>,
    to: String,
    amount: u64,
    memo: String,
) -> Result<PartialTx, ZcashError> {
    uniffi_async_export!(context, {
        let from = get_vault_address(vault)?;
        let utxos = crate::wallet::list_utxos_async(&context, from.clone()).await?;
        let (inputs, change, fee) = select_utxos(&utxos, amount, &memo)?;
        let mut outputs = vec![];
        outputs.push(Output {
            address: to,
            amount,
            memo,
        });
        outputs.push(Output {
            address: from,
            amount: change,
            memo: String::new(),
        });
        let mut tx_seed = [0u8; 32];
        OsRng.fill_bytes(&mut tx_seed);
        let partial_tx = PartialTx {
            height,
            inputs,
            outputs,
            fee,
            tx_seed: tx_seed.to_vec(),
        };

        Ok::<_, ZcashError>(partial_tx)
    })
}

pub fn combine_vault(height: u32, vault: Vec<u8>) -> Result<PartialTx, ZcashError> {
    uniffi_async_export!(context, {
        let from = get_vault_address(vault.clone())?;
        let utxos = crate::wallet::list_utxos_async(&context, from.clone()).await?;
        combine_vault_utxos_async(height, vault, utxos).await
    })
}

pub fn combine_vault_utxos(
    height: u32,
    vault: Vec<u8>,
    utxos: Vec<UTXO>,
) -> Result<PartialTx, ZcashError> {
    uniffi_async_export!(_config, {
        combine_vault_utxos_async(height, vault, utxos).await
    })
}

async fn combine_vault_utxos_async(
    height: u32,
    vault: Vec<u8>,
    utxos: Vec<UTXO>,
) -> Result<PartialTx, ZcashError> {
    let from = get_vault_address(vault)?;
    let total = utxos.iter().map(|utxo| utxo.value).sum::<u64>();
    let fee = utxos.len() as u64 * BASE_FEE;
    let amount = total - fee;
    let outputs = vec![Output {
        address: from,
        amount,
        memo: String::new(),
    }];
    let mut tx_seed = [0u8; 32];
    OsRng.fill_bytes(&mut tx_seed);
    let ptx = PartialTx {
        height,
        inputs: utxos,
        outputs,
        fee,
        tx_seed: tx_seed.to_vec(),
    };

    Ok::<_, ZcashError>(ptx)
}

fn pay_with_utxos(
    context: &Context,
    expiry_height: u32,
    sk: SecretKey,
    utxos: Vec<UTXO>,
    from_addr: String,
    to_addr: String,
    amount: u64,
    change: u64,
    memo: String,
) -> Result<TxBytes, ZcashError> {
    let network = context.config.network();
    let mut txbuilder = TxBuilder::new(
        network,
        BlockHeight::from_u32(expiry_height),
        BuildConfig::Standard {
            sapling_anchor: None,
            orchard_anchor: None,
        },
    );
    for utxo in utxos {
        let op = OutPoint::new(to_hash(&utxo.txid)?, utxo.vout);
        let coin = TxOut {
            value: Zatoshis::from_u64(utxo.value).unwrap(),
            script_pubkey: Script(hex::decode(&utxo.script).unwrap().to_vec()),
        };
        txbuilder
            .add_transparent_input(sk, op, coin)
            .map_err(to_zcasherror(anyhow!("Cannot add utxo {utxo:?}")))?;
    }
    if !memo.is_empty() {
        txbuilder
            .add_transparent_output_memo(memo.as_bytes())
            .map_err(to_zcasherror(anyhow!("Cannot add memo {memo}")))?;
    }
    let from_taddr = TransparentAddress::decode(&network, &from_addr)
        .map_err(to_zcasherror(anyhow!("Invalid source address {from_addr}")))?;
    let to_taddr = TransparentAddress::decode(&network, &to_addr).map_err(to_zcasherror(
        anyhow!("Invalid destination address {to_addr}"),
    ))?;
    txbuilder
        .add_transparent_output(&to_taddr, Zatoshis::from_u64(amount).unwrap())
        .map_err(to_zcasherror(anyhow!(
            "Cannot add output {to_addr} {amount}"
        )))?;
    txbuilder
        .add_transparent_output(&from_taddr, Zatoshis::from_u64(change).unwrap())
        .map_err(to_zcasherror(anyhow!(
            "Cannot add output {to_addr} {amount}"
        )))?;

    let prover = &context.sapling_prover;
    let res = txbuilder
        .build(OsRng, prover, prover, &zip317::FeeRule::standard())
        .map_err(|e| ZcashError::AssertError(e.to_string()))?;

    let tx = res.transaction();
    let txid = tx.txid().to_string();
    let mut bytes = vec![];
    tx.write(&mut bytes).unwrap();
    let tx = TxBytes { txid, data: bytes };

    Ok(tx)
}

pub struct Sighashes {
    pub hashes: Vec<Vec<u8>>,
}

pub fn build_vault_unauthorized_tx(
    vault: Vec<u8>,
    ptx: PartialTx,
) -> Result<Sighashes, ZcashError> {
    uniffi_export!(context, {
        let network = context.config.network();
        let mut tx_rng = rand_chacha::ChaCha20Rng::from_seed(to_ba(&ptx.tx_seed)?);
        let pk = PublicKey::from_slice(&vault).map_err(|_| ZcashError::InvalidVaultPubkey)?;
        let ovk = get_ovk(vault)?;

        let mut tbuilder = TransparentBuilder::empty();
        for i in ptx.inputs.iter() {
            let UTXO {
                txid,
                vout,
                script,
                value,
                ..
            } = i;
            let op = OutPoint::new(to_hash(&txid)?, *vout);
            let coin = TxOut {
                value: Zatoshis::from_u64(*value).unwrap(),
                script_pubkey: Script(hex::decode(script).unwrap()),
            };
            tbuilder
                .add_input_without_sk(pk, op, coin)
                .map_err(|e| ZcashError::AssertError(e.to_string()))?;
        }

        let mut sbuilder = sapling_crypto::builder::Builder::new(
            Zip212Enforcement::On,
            sapling_crypto::builder::BundleType::Transactional {
                bundle_required: false,
            },
            Anchor::empty_tree(), // not required when there are no shielded input
        );
        let mut obuilder = orchard::builder::Builder::new(
            BundleType::Transactional {
                flags: Flags::ENABLED,
                bundle_required: false,
            },
            orchard::Anchor::empty_tree(),
        );

        for o in ptx.outputs.iter() {
            let Output {
                address,
                amount,
                memo,
            } = o;
            let recipient = Address::decode(&network, &address)
                .ok_or(ZcashError::InvalidAddress(address.clone()))?;

            let mut hr = |receiver: Receiver| {
                handle_receiver(
                    receiver,
                    *amount,
                    &memo,
                    &ovk,
                    &mut tbuilder,
                    &mut sbuilder,
                    &mut obuilder,
                )
            };

            match recipient {
                Address::Tex(_) => Err(ZcashError::InvalidAddress(address.clone())),
                Address::Transparent(transparent_address) => {
                    hr(Receiver::Transparent(transparent_address))
                }
                Address::Sapling(payment_address) => hr(Receiver::Sapling(payment_address)),
                Address::Unified(unified_address) => {
                    if let Some(&receiver) = unified_address.orchard() {
                        hr(Receiver::Orchard(receiver))
                    } else if let Some(&receiver) = unified_address.sapling() {
                        hr(Receiver::Sapling(receiver))
                    } else if let Some(&receiver) = unified_address.transparent() {
                        hr(Receiver::Transparent(receiver))
                    } else {
                        Err(ZcashError::AssertError("Unreachable".into()))
                    }
                }
            }?;
        }

        let tbundle = tbuilder.build();
        let sbundle = sbuilder
            .build::<LocalTxProver, LocalTxProver, _, Amount>(&mut tx_rng)
            .map_err(to_zcasherror(anyhow!("Cannot build sapling bundle")))?
            .map(|(bundle, _)| {
                let prover: &LocalTxProver = &context.sapling_prover;
                bundle.create_proofs(prover, prover, &mut tx_rng, ())
            });
        let obundle = obuilder
            .build::<Amount>(&mut tx_rng)
            .map_err(to_zcasherror(anyhow!("Cannot build orchard bundle")))?
            .map(|v| v.0);

        let height = BlockHeight::from_u32(ptx.height);
        let consensus_branch_id = BranchId::for_height(&network, height);
        let version = TxVersion::suggested_for_branch(consensus_branch_id);
        let unauthed_tx: TransactionData<zcash_primitives::transaction::Unauthorized> =
            TransactionData::from_parts(
                version,
                consensus_branch_id,
                0,
                height,
                tbundle,
                None,
                sbundle,
                obundle,
            );
        let txid_parts = unauthed_tx.digest(TxIdDigester);
        let _txid = signature_hash(&unauthed_tx, &SignableInput::Shielded, &txid_parts)
            .as_ref()
            .to_vec();

        let mut sighashes = vec![];
        for (index, inp) in ptx.inputs.iter().enumerate() {
            let script = Script(hex::decode(&inp.script).unwrap());
            let sighash = signature_hash(
                &unauthed_tx,
                &SignableInput::Transparent {
                    hash_type: SIGHASH_ALL,
                    index,
                    script_code: &script,
                    script_pubkey: &script,
                    value: Zatoshis::from_u64(inp.value).unwrap(),
                },
                &txid_parts,
            )
            .as_ref()
            .to_vec();
            sighashes.push(sighash);
        }
        let sighashes = Sighashes { hashes: sighashes };
        Ok::<_, ZcashError>(sighashes)
    })
}

fn handle_receiver(
    receiver: Receiver,
    amount: u64,
    memo: &str,
    ovk: &[u8],
    tbuilder: &mut TransparentBuilder,
    sbuilder: &mut sapling_crypto::builder::Builder,
    obuilder: &mut orchard::builder::Builder,
) -> Result<(), ZcashError> {
    let memo_bytes = if memo.is_empty() {
        None
    } else {
        let memo =
            Memo::from_str(&memo).map_err(|_| ZcashError::AssertError("Invalid memo".into()))?;
        let memo = MemoBytes::try_from(memo).unwrap();
        Some(memo.as_array().clone())
    };
    match receiver {
        Receiver::Transparent(transparent_address) => {
            let amount = Zatoshis::from_u64(amount).unwrap();
            tbuilder
                .add_output(&transparent_address, amount)
                .map_err(to_zcasherror(anyhow!("Cannot add transparent output")))?;
            if !memo.is_empty() {
                tbuilder
                    .add_output_memo(memo.as_bytes())
                    .map_err(to_zcasherror(anyhow!("Cannot add transparent memo")))?;
            }
        }
        Receiver::Sapling(payment_address) => {
            sbuilder
                .add_output(
                    Some(sapling_crypto::keys::OutgoingViewingKey(to_ba(ovk)?)),
                    payment_address,
                    sapling_crypto::value::NoteValue::from_raw(amount),
                    memo_bytes,
                )
                .map_err(|e| ZcashError::AssertError(e.to_string()))?;
        }
        Receiver::Orchard(address) => {
            obuilder
                .add_output(
                    Some(OutgoingViewingKey::from(to_ba(ovk)?)),
                    address,
                    NoteValue::from_raw(amount),
                    memo_bytes,
                )
                .map_err(|e| ZcashError::AssertError(e.to_string()))?;
        }
    }

    Ok::<_, ZcashError>(())
}
