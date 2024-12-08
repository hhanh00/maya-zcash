use std::{
    cmp::{max, min},
    path::PathBuf,
};

use anyhow::anyhow;
use rand_core::OsRng;
use secp256k1::SecretKey;
use zcash_keys::encoding::AddressCodec;
use zcash_primitives::{
    legacy::{Script, TransparentAddress},
    transaction::{
        builder::{BuildConfig, Builder as TxBuilder},
        components::{OutPoint, TxOut},
        fees::zip317,
    },
};
use zcash_proofs::prover::LocalTxProver;
use zcash_protocol::{consensus::BlockHeight, value::Zatoshis};

use crate::{
    addr::{get_vault_address, validate_address},
    config::Config,
    to_hash, to_zcasherror, uniffi_async_export,
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
    uniffi_async_export!(config, {
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

        let utxos = crate::wallet::list_utxos_async(config, from.clone()).await?;
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
            inputs.push(utxo);

            if needed == 0 {
                break;
            }
        }
        if needed > 0 {
            return Err(ZcashError::NotEnoughFunds);
        }
        let f = fee(num_tins, num_touts);
        let change = input_amount - amount - f;

        let txb = pay_with_utxos(
            config,
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

fn pay_with_utxos(
    config: &Config,
    expiry_height: u32,
    sk: SecretKey,
    utxos: Vec<UTXO>,
    from_addr: String,
    to_addr: String,
    amount: u64,
    change: u64,
    memo: String,
) -> Result<TxBytes, ZcashError> {
    let network = config.network();
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
    let from_taddr = TransparentAddress::decode(&config.network(), &from_addr)
        .map_err(to_zcasherror(anyhow!("Invalid source address {from_addr}")))?;
    let to_taddr = TransparentAddress::decode(&config.network(), &to_addr).map_err(
        to_zcasherror(anyhow!("Invalid destination address {to_addr}")),
    )?;
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

    let param_dir = PathBuf::from(&config.sapling_params_dir);
    let prover = LocalTxProver::new(
        param_dir.join("sapling-spend.params").as_path(),
        param_dir.join("sapling-output.params").as_path(),
    );

    let res = txbuilder
        .build(OsRng, &prover, &prover, &zip317::FeeRule::standard())
        .map_err(|e| ZcashError::AssertError(e.to_string()))?;

    let tx = res.transaction();
    let txid = tx.txid().to_string();
    let mut bytes = vec![];
    tx.write(&mut bytes).unwrap();
    let tx = TxBytes { txid, data: bytes };

    Ok(tx)
}
