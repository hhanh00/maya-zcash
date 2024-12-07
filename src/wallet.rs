use anyhow::Context as _;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::{rpc::{json_request, map_rpc_error}, uniffi_async_export, ZcashError};

pub fn get_balance(address: String) -> Result<u64, ZcashError> {
    uniffi_async_export!(config, {
        let id = Uuid::new_v4().to_string();

        let rep = json_request(config, &id, "getaddressbalance",
            vec![address.into()]).await
            .map_err(map_rpc_error)?;
        let balance = rep["balance"].as_u64().ok_or(ZcashError::RPC("No balance field".to_string()))?;

        Ok(balance)
    })
}

#[derive(Serialize, Deserialize)]
pub struct UTXO {
    pub txid: String,
    pub height: u32,
    #[serde(rename = "outputIndex")]
    pub vout: u32,
    pub script: String,
    #[serde(rename = "satoshis")]
    pub value: u64,
}

pub fn list_utxos(address: String) -> Result<Vec<UTXO>, ZcashError> {
    uniffi_async_export!(config, {
        let id = Uuid::new_v4().to_string();

        let rep = json_request(config, &id, "getaddressutxos",
            vec![address.into()]).await
            .map_err(map_rpc_error)?;
        let list_utxos: Vec<UTXO> = serde_json::from_value(rep)
        .context("Could not parse getaddressutxos response")
        .map_err(map_rpc_error)?;

        Ok(list_utxos)
    })
}
