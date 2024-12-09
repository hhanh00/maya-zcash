use anyhow::Context as _;
use base58check::FromBase58Check as _;
use secp256k1::{All, PublicKey, Secp256k1, SecretKey};
use serde::{Deserialize, Serialize};
use sha2::Digest as _;
use uuid::Uuid;
use zcash_primitives::legacy::TransparentAddress;

use crate::{
    config::Context,
    rpc::{json_request, map_rpc_error},
    uniffi_async_export, uniffi_export, ZcashError,
};

pub fn get_balance(address: String) -> Result<u64, ZcashError> {
    uniffi_async_export!(context, {
        let config = &context.config;
        let id = Uuid::new_v4().to_string();
        let rep = json_request(config, &id, "getaddressbalance", vec![address.into()])
            .await
            .map_err(map_rpc_error)?;
        let balance = rep["balance"]
            .as_u64()
            .ok_or(ZcashError::RPC("No balance field".to_string()))?;

        Ok(balance)
    })
}

#[derive(Clone, Serialize, Deserialize, Debug)]
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
    uniffi_async_export!(context, { list_utxos_async(&context, address).await })
}

pub async fn list_utxos_async(context: &Context, address: String) -> Result<Vec<UTXO>, ZcashError> {
    let config = &context.config;
    let id = Uuid::new_v4().to_string();

    let rep = json_request(config, &id, "getaddressutxos", vec![address.into()])
        .await
        .map_err(map_rpc_error)?;
    let list_utxos: Vec<UTXO> = serde_json::from_value(rep)
        .context("Could not parse getaddressutxos response")
        .map_err(map_rpc_error)?;

    Ok(list_utxos)
}

pub struct TransparentKey {
    pub sk: Vec<u8>,
    pub pk: Vec<u8>,
    pub addr: String,
}

pub fn sk_to_pub(wif: String) -> Result<TransparentKey, ZcashError> {
    uniffi_export!(context, {
        let network = context.config.network();
        let (_, sk) = wif
            .from_base58check()
            .map_err(|_| anyhow::anyhow!("Not Base58 Encoded"))?;
        let skb = &sk[0..sk.len() - 1]; // remove compressed pub key marker
        let sk = SecretKey::from_slice(&skb)
            .context("Cannot parse secret key")
            .map_err(ZcashError::from)?;
        let secp = Secp256k1::<All>::new();
        let pk = PublicKey::from_secret_key(&secp, &sk);
        let pk = pk.serialize().to_vec();
        let sha = sha2::Sha256::digest(&pk);
        let pkh: [u8; 20] = ripemd::Ripemd160::digest(&sha).into();
        let addr = TransparentAddress::PublicKeyHash(pkh);
        let addr = zcash_client_backend::address::Address::Transparent(addr);
        let addr = addr.encode(&network);
        let tk = TransparentKey {
            sk: skb.to_vec(),
            pk,
            addr,
        };
        Ok(tk)
    })
}
