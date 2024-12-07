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
