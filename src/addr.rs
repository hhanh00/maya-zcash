use anyhow::Result;
use sha2::Digest as _;
use zcash_keys::address::Address;
use zcash_primitives::legacy::TransparentAddress;

use crate::{uniffi_export, ZcashError};

pub fn get_vault_address(pubkey: Vec<u8>) -> Result<String, ZcashError> {
    if pubkey.len() != 33 {
        return Err(ZcashError::InvalidPubkeyLength(hex::encode(&pubkey)));
    }
    let taddr = uniffi_export!(config, {
        let sha = sha2::Sha256::digest(&pubkey);
        let pkh: [u8; 20] = ripemd::Ripemd160::digest(&sha).into();
        let tkey = TransparentAddress::PublicKeyHash(pkh);
        let taddr = zcash_client_backend::address::Address::Transparent(tkey);
        let taddr = taddr.encode(&config.network());
        taddr
    });
    Ok(taddr)
}

pub fn validate_address(address: String) -> Result<bool, ZcashError> {
    uniffi_export!(config, {
        let r = Address::decode(&config.network(), &address);
        Ok(r.is_some())
    })
}
