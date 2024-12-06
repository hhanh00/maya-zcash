use anyhow::Result;
use uuid::Uuid;

use crate::{rpc::json_request, uniffi_async_export, Height, ZcashError, CONTEXT};

pub fn get_latest_height() -> Result<Height, ZcashError> {
    uniffi_async_export!(config, {
        let id = Uuid::new_v4().to_string();

        let result = json_request(config, &id, "getblockcount", vec![]).await
        .map_err(|e| ZcashError::RPC(e.to_string()))?;
        let height = result.as_i64().unwrap();

        let result = json_request(config, &id, "getblockhash", vec![height.into()]).await
        .map_err(|e| ZcashError::RPC(e.to_string()))?;
        let hash = result.as_str().unwrap().to_string();

        Height {
            number: height as u32,
            hash: hex::decode(&hash).unwrap(),
        }
    })
}

