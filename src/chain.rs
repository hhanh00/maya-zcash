use anyhow::Result;
use uuid::Uuid;

use crate::{data_generated::fb::HeightT, rpc::json_request, uniffi_export, ZcashError, CONTEXT};

pub fn get_latest_height() -> Result<Vec<u8>, ZcashError> {
    uniffi_export!(config, HeightT, {
        let id = Uuid::new_v4().to_string();

        let result = json_request(config, &id, "getblockcount", vec![]).await
        .map_err(|e| ZcashError::RPC(e.to_string()))?;
        let height = result.as_i64().unwrap();

        let result = json_request(config, &id, "getblockhash", vec![height.into()]).await
        .map_err(|e| ZcashError::RPC(e.to_string()))?;
        let hash = result.as_str().unwrap().to_string();

        HeightT {
            number: height as u32,
            hash: hex::decode(&hash).ok(),
        }
    })
}

