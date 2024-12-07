use anyhow::Result;
use uuid::Uuid;

use crate::{rpc::{json_request, map_rpc_error}, uniffi_async_export, Height, ZcashError};

pub fn get_latest_height() -> Result<Height, ZcashError> {
    uniffi_async_export!(config, {
        let id = Uuid::new_v4().to_string();

        let result = json_request(config, &id, "getblockcount", vec![]).await
        .map_err(map_rpc_error)?;
        let height = result.as_i64().unwrap();

        let result = json_request(config, &id, "getblockhash", vec![height.into()]).await
        .map_err(map_rpc_error)?;
        let hash = result.as_str().unwrap().to_string();

        Ok(Height {
            number: height as u32,
            hash: hex::decode(&hash).unwrap(),
        })
    })
}

