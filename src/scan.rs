use anyhow::Context as _;
use serde::{Deserialize, Serialize};
use serde_json::json;
use uuid::Uuid;

use crate::{
    addr::{get_ovk, get_vault_address},
    rpc::{json_request, map_rpc_error},
    uniffi_async_export, ZcashError,
};

pub struct Note {
    pub address: String,
    pub value: i64,
    pub memo: Option<String>,
}

pub struct TxData {
    pub txid: String,
    pub height: u32, // block height or 0 if unconfirmed
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
pub struct TIn {}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct TOut {}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct SIn {}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct SOut {}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct Orchard {
    pub actions: Vec<Action>,
}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct Action {}

#[derive(Clone, Serialize, Deserialize, Debug)]
pub struct MempoolTx {
    #[serde(rename = "vin")]
    tins: Vec<TIn>,
    #[serde(rename = "vout")]
    touts: Vec<TOut>,
    #[serde(rename = "vShieldedSpend")]
    sins: Vec<SIn>,
    #[serde(rename = "vShieldedOutput")]
    souts: Vec<SOut>,
    orchard: Orchard,
}

pub fn scan_mempool(pubkey: Vec<u8>) -> Result<Vec<TxData>, ZcashError> {
    uniffi_async_export!(config, {
        let vault_addr = get_vault_address(pubkey.clone())?;
        let _ovk = get_ovk(pubkey)?;

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

        for tx in delta.iter() {
            let id = Uuid::new_v4().to_string();
            let rep = json_request(
                config,
                &id,
                "getrawtransaction",
                vec![tx.txid.clone().into(), 1.into()],
            )
            .await
            .map_err(map_rpc_error)?;
            tracing::info!("{:?}", rep);
            let tx: MempoolTx = serde_json::from_value(rep)
                .context("Cannot parse getrawtransaction reply")
                .map_err(map_rpc_error)?;
            tracing::info!("{:?}", tx);
        }

        Ok(vec![])
    })
    // pubkey -> taddr, ovk
    // for each tx
    // - decode into t/z/o bundles
    // - check tbundle
    //   - i/o not contains taddr -> continue
    //   - resolve inp (txid, vout) -> (addr, value)
    //   - decode OP_RESULT if any
    // - check z/o bundles
    //   - try decrypt out with ovk
    //   - if tinp contains taddr, every output should be decryptable
    //     (addr, value)
    //   - inp are always unknown
}
