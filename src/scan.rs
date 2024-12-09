use std::collections::HashSet;

use anyhow::{anyhow, Context as _};
use orchard::{
    note_encryption::OrchardDomain,
    primitives::redpallas::{SpendAuth, VerificationKey},
};
use sapling_crypto::{
    bundle::OutputDescription, note::ExtractedNoteCommitment, note_encryption::SaplingDomain,
    value::ValueCommitment,
};
use serde::{Deserialize, Serialize};
use serde_json::json;
use uuid::Uuid;
use zcash_keys::{address::UnifiedAddress, encoding::AddressCodec};
use zcash_note_encryption::{EphemeralKeyBytes, ENC_CIPHERTEXT_SIZE, OUT_CIPHERTEXT_SIZE};
use zcash_protocol::memo::{Memo, MemoBytes};

use crate::{
    addr::{get_ovk, get_vault_address},
    config::{Config, Context},
    network::Network,
    pay::Output,
    rpc::{json_request, map_rpc_error},
    to_ba, to_hash, to_uhash, to_zcasherror, uniffi_async_export, ZcashError,
};

pub struct Note {
    pub address: String,
    pub value: i64,
    pub memo: Option<String>,
}

pub struct TxData {
    pub txid: String,
    // Who paid us or who we are paying, and how much
    // value > 0 -> vault received funds
    pub counterparty: Note,
    // All the transparent input/output and the shielded outputs
    // we can decrypt
    pub plain: Vec<Note>,
    // amount encrypted
    pub encrypted: i64,
    pub fee: u64,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct MempoolTxDelta {
    address: String,
    txid: String,
    #[serde(rename = "index")]
    vout: u32,
    #[serde(rename = "satoshis")]
    value: i64,
    timestamp: u32,
    prevtxid: Option<String>,
    prevout: Option<u32>,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct TIn {
    txid: String,
    vout: u32,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct ScriptPubKey {
    addresses: Option<Vec<String>>,
    asm: String,
    r#type: String,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct TRawOut {
    #[serde(rename = "n")]
    pub vout: u32,
    #[serde(rename = "scriptPubKey")]
    pub script: ScriptPubKey,
    #[serde(rename = "valueSat")]
    pub value: u64,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct TOut {
    pub address: String,
    pub value: u64,
    pub memo: Option<String>,
}

impl TryFrom<TRawOut> for TOut {
    type Error = ();

    fn try_from(out: TRawOut) -> Result<Self, Self::Error> {
        match out.script.r#type.as_str() {
            "pubkeyhash" => Ok(TOut {
                address: out.script.addresses.unwrap()[0].clone(),
                value: out.value,
                memo: None,
            }),
            "nulldata" => {
                let asm = &out.script.asm;
                if asm.starts_with("OP_RETURN") {
                    let memo = hex::decode(&asm[10..]).unwrap();
                    let memo = String::from_utf8_lossy(&memo).to_string();
                    Ok(TOut {
                        address: String::new(),
                        value: 0,
                        memo: Some(memo),
                    })
                } else {
                    Err(())
                }
            }
            _ => Err(()),
        }
    }
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct SOut {
    pub cv: String,
    pub cmu: String,
    #[serde(rename = "ephemeralKey")]
    pub epk: String,
    #[serde(rename = "encCiphertext")]
    pub enc: String,
    #[serde(rename = "outCiphertext")]
    pub out: String,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct Orchard {
    pub actions: Vec<Action>,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct Action {
    pub cv: String,
    pub cmx: String,
    #[serde(rename = "nullifier")]
    pub rho: String,
    pub rk: String,
    #[serde(rename = "ephemeralKey")]
    pub epk: String,
    #[serde(rename = "encCiphertext")]
    pub enc: String,
    #[serde(rename = "outCiphertext")]
    pub out: String,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct RawVaultTx {
    txid: String,
    #[serde(rename = "vin")]
    tins: Vec<TIn>,
    #[serde(rename = "vout")]
    touts: Vec<TRawOut>,
    #[serde(rename = "vShieldedOutput")]
    souts: Vec<SOut>,
    orchard: Orchard,
}

#[derive(Clone, Debug)]
pub struct VaultTxDetails {
    txid: String,
    tins: Vec<TIn>,
    ptouts: Vec<TOut>,
    touts: Vec<TOut>,
    souts: Vec<SOut>,
    actions: Vec<Action>,
}

#[derive(Debug)]
pub enum Direction {
    Incoming,
    Outgoing,
}

#[derive(Debug)]
pub struct VaultTx {
    pub height: u32,
    pub txid: String,
    pub counterparty: Output,
    pub direction: Direction,
}

impl From<RawVaultTx> for VaultTxDetails {
    fn from(tx: RawVaultTx) -> Self {
        Self {
            txid: tx.txid.clone(),
            tins: tx.tins,
            ptouts: vec![],
            touts: tx
                .touts
                .into_iter()
                .filter_map(|o| o.try_into().ok())
                .collect(),
            souts: tx.souts,
            actions: tx.orchard.actions,
        }
    }
}

pub struct VaultTxDecrypted {
    txid: String,
    ptouts: Vec<TOut>,
    outputs: Vec<Output>,
}

impl VaultTxDetails {
    pub async fn resolve_inputs(&mut self, config: &Config) -> Result<(), ZcashError> {
        for tin in self.tins.iter() {
            let id = Uuid::new_v4().to_string();
            let rep = json_request(
                config,
                &id,
                "getrawtransaction",
                vec![tin.txid.clone().into(), 1.into()],
            )
            .await
            .map_err(map_rpc_error)?;
            let tx: RawVaultTx = serde_json::from_value(rep)
                .context("Cannot parse getrawtransaction reply")
                .map_err(map_rpc_error)?;
            self.ptouts
                .push(tx.touts[tin.vout as usize].clone().try_into().unwrap());
        }
        Ok(())
    }

    pub fn decrypt(
        &self,
        network: &Network,
        ovk: [u8; 32],
    ) -> Result<VaultTxDecrypted, ZcashError> {
        let mut outputs = vec![];

        let mut tmemo = None;
        for tout in self.touts.iter() {
            if tout.memo.is_some() {
                tmemo = tout.memo.clone();
            }
            outputs.push(Output {
                address: tout.address.clone(),
                amount: tout.value,
                memo: String::new(),
            });
        }
        if let Some(tmemo) = tmemo {
            for o in outputs.iter_mut() {
                o.memo = tmemo.clone();
            }
        }

        for sout in self.souts.iter() {
            let cv = to_hash(&sout.cv)?;
            let cv = ValueCommitment::from_bytes_not_small_order(&cv).unwrap();
            let cmu = to_hash(&sout.cmu)?;
            let cmu = ExtractedNoteCommitment::from_bytes(&cmu).unwrap();
            let epk = to_hash(&sout.epk)?;
            let epk = EphemeralKeyBytes::from(epk);
            let mut enc = [0u8; ENC_CIPHERTEXT_SIZE];
            enc.copy_from_slice(&hex::decode(&sout.enc).unwrap());
            let mut out = [0u8; OUT_CIPHERTEXT_SIZE];
            out.copy_from_slice(&hex::decode(&sout.out).unwrap());
            let output = OutputDescription::<()>::from_parts(cv, cmu, epk, enc, out, ());
            let d = SaplingDomain::new(sapling_crypto::note_encryption::Zip212Enforcement::On);
            let ovk = sapling_crypto::keys::OutgoingViewingKey(ovk);
            if let Some((note, address, memo)) = zcash_note_encryption::try_output_recovery_with_ovk(
                &d,
                &ovk,
                &output,
                &output.cv(),
                &out,
            ) {
                let memo = memo_to_string(&memo);
                outputs.push(Output {
                    address: address.encode(network),
                    amount: note.value().inner(),
                    memo,
                });
            }
        }

        for a in self.actions.iter() {
            let rho = to_uhash(&a.rho)?;
            let rho = orchard::note::Nullifier::from_bytes(&rho).unwrap();
            let rk = to_uhash(&a.rk)?;
            let rk = VerificationKey::<SpendAuth>::try_from(rk).unwrap();
            let cv = to_uhash(&a.cv)?;
            let cv = orchard::value::ValueCommitment::from_bytes(&cv).unwrap();
            let cmx = to_uhash(&a.cmx)?;
            let cmx = orchard::note::ExtractedNoteCommitment::from_bytes(&cmx).unwrap();
            let epk = to_uhash(&a.epk)?;
            let mut enc = [0u8; ENC_CIPHERTEXT_SIZE];
            enc.copy_from_slice(&hex::decode(&a.enc).unwrap());
            let mut out = [0u8; OUT_CIPHERTEXT_SIZE];
            out.copy_from_slice(&hex::decode(&a.out).unwrap());
            let encrypted_note = orchard::note::TransmittedNoteCiphertext {
                epk_bytes: epk,
                enc_ciphertext: enc,
                out_ciphertext: out,
            };
            let action = orchard::Action::from_parts(rho, rk, cmx, encrypted_note, cv, ());
            let d = OrchardDomain::for_action(&action);
            let ovk = orchard::keys::OutgoingViewingKey::from(ovk);
            if let Some((note, address, memo)) = zcash_note_encryption::try_output_recovery_with_ovk(
                &d,
                &ovk,
                &action,
                action.cv_net(),
                &action.encrypted_note().out_ciphertext,
            ) {
                let address = UnifiedAddress::from_receivers(Some(address), None, None).unwrap();
                let memo = memo_to_string(&memo);
                outputs.push(Output {
                    address: address.encode(network),
                    amount: note.value().inner(),
                    memo,
                });
            }
        }

        let txdec = VaultTxDecrypted {
            txid: self.txid.clone(),
            ptouts: self.ptouts.clone(),
            outputs,
        };

        Ok(txdec)
    }
}

impl VaultTx {
    fn from_decrypted(
        height: u32,
        txd: &VaultTxDecrypted,
        vault_addr: &str,
    ) -> Result<Self, ZcashError> {
        let spent = txd
            .ptouts
            .iter()
            .filter_map(|pout| {
                if pout.address == vault_addr {
                    Some(pout.value)
                } else {
                    None
                }
            })
            .sum::<u64>();
        let tx_vault = if spent > 0 {
            // outgoing
            let non_vault_outputs = txd
                .outputs
                .iter()
                .filter(|&o| o.address != vault_addr && o.amount > 0)
                .cloned()
                .collect::<Vec<_>>();
            if non_vault_outputs.len() != 1 {
                return Err(ZcashError::AssertError(
                    "Payment from vault should have a single recipient".into(),
                ));
            }
            VaultTx {
                height,
                txid: txd.txid.clone(),
                counterparty: non_vault_outputs.first().unwrap().clone(),
                direction: Direction::Outgoing,
            }
        } else {
            // spent is 0, there are no vault inputs, which means
            // there must be vault outputs
            let vault_outputs = txd
                .outputs
                .iter()
                .filter(|&o| o.address == vault_addr && o.amount > 0)
                .cloned()
                .collect::<Vec<_>>();
            if vault_outputs.is_empty() {
                return Err(ZcashError::AssertError(
                    "Payment to vault should have a vault output".into(),
                ));
            }
            let total_value = vault_outputs.iter().map(|o| o.amount).sum::<u64>();
            // there can be only at most one transparent memo per tx
            let memo = vault_outputs.first().unwrap().memo.clone();
            let mut counterparty_addr = String::new();
            // if the deposit/swap into the vault came from multiple transparent
            // sources, pick the first one
            // shielded sources are unknown
            // the counterparty_addr may remain unknown
            if let Some(first_tin) = txd.ptouts.first() {
                counterparty_addr = first_tin.address.clone();
            }
            VaultTx {
                height,
                txid: txd.txid.clone(),
                counterparty: Output {
                    address: counterparty_addr,
                    amount: total_value,
                    memo,
                },
                direction: Direction::Incoming,
            }
        };

        Ok(tx_vault)
    }
}

pub fn scan_mempool(pubkey: Vec<u8>) -> Result<Vec<VaultTx>, ZcashError> {
    uniffi_async_export!(context, {
        let config = &context.config;
        let vault_addr = get_vault_address(pubkey.clone())?;
        let ovk = get_ovk(pubkey)?;

        let id = Uuid::new_v4().to_string();
        let rep = json_request(
            config,
            &id,
            "getaddressmempool",
            vec![json!({
                "addresses": [vault_addr]
            })],
        )
        .await
        .map_err(map_rpc_error)?;
        let delta: Vec<MempoolTxDelta> = serde_json::from_value(rep)
            .context("Cannot parse getaddressmempool reply")
            .map_err(map_rpc_error)?;
        let tx_ids: HashSet<String> = delta.into_iter().map(|d| d.txid).collect();

        let mut txs = vec![];
        for txid in tx_ids.iter() {
            if let Some(tx) = process_tx(
                config,
                &TxId {
                    txid: txid.clone(),
                    height: 0,
                },
                &vault_addr,
                &ovk,
            )
            .await?
            {
                txs.push(tx);
            }
        }

        Ok(txs)
    })
}

async fn process_tx(
    config: &Config,
    txid: &TxId,
    vault_addr: &str,
    ovk: &[u8],
) -> Result<Option<VaultTx>, ZcashError> {
    let network = config.network();

    let id = Uuid::new_v4().to_string();
    let rep = json_request(
        config,
        &id,
        "getrawtransaction",
        vec![txid.txid.clone().into(), 1.into()],
    )
    .await
    .map_err(map_rpc_error)?;
    tracing::debug!("{:?}", rep);

    let tx: RawVaultTx = serde_json::from_value(rep)
        .context("Cannot parse getrawtransaction reply")
        .map_err(map_rpc_error)?;
    let mut tx: VaultTxDetails = tx.into();
    if tx.touts.iter().any(|o| o.address == vault_addr) {
        tx.resolve_inputs(&config).await?;
        let txd = tx.decrypt(&network, to_ba(&ovk)?)?;
        let tx = VaultTx::from_decrypted(txid.height, &txd, &vault_addr)?;
        tracing::info!("{:?}", tx);
        return Ok(Some(tx));
    }
    Ok(None)
}

pub struct BlockTxs {
    pub start_hash: String,
    pub end_hash: String,
    pub start_height: u32,
    pub end_height: u32,
    pub txs: Vec<VaultTx>,
}

#[derive(Serialize, Deserialize, Debug)]
struct BlockHeader {
    hash: String,
    height: u32,
    #[serde(rename = "previousblockhash")]
    prev_hash: String,
    #[serde(rename = "nextblockhash")]
    next_hash: Option<String>,
}

#[derive(Serialize, Deserialize)]
struct TxId {
    txid: String,
    height: u32,
}

#[derive(Serialize, Deserialize)]
struct BlockHeight {
    hash: String,
    height: u32,
}

#[derive(Serialize, Deserialize)]
struct AddressDeltas {
    #[serde(rename = "deltas")]
    txids: Vec<TxId>,
    start: BlockHeight,
    end: BlockHeight,
}

pub fn scan_blocks(
    pubkey: Vec<u8>,
    mut prev_hashes: Vec<String>,
) -> Result<Option<BlockTxs>, ZcashError> {
    uniffi_async_export!(context, {
        let config = &context.config;
        let vault_addr = get_vault_address(pubkey.clone())?;
        let ovk = get_ovk(pubkey)?;

        if prev_hashes.is_empty() {
            let id = Uuid::new_v4().to_string();
            let rep = json_request(config, &id, "getblockhash", vec![1.into()])
                .await
                .map_err(map_rpc_error)?;
            let genesis_hash = rep
                .as_str()
                .ok_or(ZcashError::AssertError(
                    "Failed to retrieve genesis hash".into(),
                ))?
                .to_string();
            prev_hashes = vec![genesis_hash];
        }

        for prev_hash in prev_hashes {
            if let Ok(block_txs) = scan_blocks_async(&context, &vault_addr, &ovk, prev_hash).await {
                return Ok(block_txs);
            }
        }

        Err(ZcashError::AssertError("Not reachable".into()))
    })
}

async fn scan_blocks_async(
    context: &Context,
    vault_addr: &str,
    ovk: &[u8],
    prev_hash: String,
) -> Result<Option<BlockTxs>, ZcashError> {
    let config = &context.config;
    let id = Uuid::new_v4().to_string();
    let rep = json_request(config, &id, "getblockheader", vec![prev_hash.into()])
        .await
        .map_err(map_rpc_error)?;
    let block_header: BlockHeader = serde_json::from_value(rep).map_err(to_zcasherror(anyhow!(
        "Failed to parse getblockheader reply"
    )))?;
    let Some(next_hash) = block_header.next_hash else {
        return Ok(None);
    };
    let start_height = block_header.height + 1;

    let id = Uuid::new_v4().to_string();
    let rep = json_request(
        config,
        &id,
        "getaddressdeltas",
        vec![json!({
            "addresses": [ vault_addr ],
            "start": start_height,
            "end": 0,
            "chainInfo": true
        })],
    )
    .await
    .map_err(map_rpc_error)?;
    let deltas: AddressDeltas = serde_json::from_value(rep).map_err(to_zcasherror(anyhow!(
        "Failed to parse getaddressdeltas reply"
    )))?;

    if deltas.start.hash != next_hash {
        return Err(ZcashError::Reorg);
    }

    for txid in deltas.txids {
        tracing::info!(">> {} {}", txid.height, txid.txid);
        process_tx(config, &txid, &vault_addr, &ovk).await?;
    }
    let btxs = BlockTxs {
        start_hash: deltas.start.hash,
        end_hash: deltas.end.hash,
        start_height: deltas.start.height,
        end_height: deltas.end.height,
        txs: vec![],
    };
    Ok(Some(btxs))
}

fn memo_to_string(memo: &[u8]) -> String {
    let memo = MemoBytes::from_bytes(memo).unwrap();
    let memo: Memo = memo.try_into().unwrap();
    let memo = match memo {
        Memo::Text(memo) => memo.to_string(),
        _ => String::new(),
    };
    memo
}
