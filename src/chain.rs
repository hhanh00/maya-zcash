use uuid::Uuid;

use crate::{
    rpc::{json_request, map_rpc_error},
    uniffi_async_export, Height, ZcashError,
};

pub fn get_latest_height() -> Result<Height, ZcashError> {
    uniffi_async_export!(context, {
        let config = &context.config;
        let id = Uuid::new_v4().to_string();
        let result = json_request(config, &id, "getblockcount", vec![])
            .await
            .map_err(map_rpc_error)?;
        let height = result.as_i64().unwrap();

        let id = Uuid::new_v4().to_string();
        let result = json_request(config, &id, "getblockhash", vec![height.into()])
            .await
            .map_err(map_rpc_error)?;
        let hash = result.as_str().unwrap().to_string();

        Ok(Height {
            number: height as u32,
            hash: hex::decode(&hash).unwrap(),
        })
    })
}

pub fn broadcast_raw_tx(txb: Vec<u8>) -> Result<String, ZcashError> {
    uniffi_async_export!(context, {
        let config = &context.config;
        let id = Uuid::new_v4().to_string();
        let tx = hex::encode(txb);
        let result = json_request(config, &id, "sendrawtransaction", vec![tx.into()])
            .await
            .map_err(map_rpc_error)?;
        let txid = result.as_str().ok_or(ZcashError::TxRejected)?.to_string();
        Ok(txid)
    })
}
