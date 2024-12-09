use crate::network::Network;
use secp256k1::PublicKey;
use sha2::Digest as _;
use zcash_address::unified::{self, Container, Encoding as _, Receiver};
use zcash_keys::address::Address;
use zcash_primitives::legacy::TransparentAddress;

use crate::{uniffi_export, ZcashError};

pub fn get_vault_address(pubkey: Vec<u8>) -> Result<String, ZcashError> {
    let _ = PublicKey::from_slice(&pubkey).map_err(|_| ZcashError::InvalidVaultPubkey)?;
    let taddr = uniffi_export!(context, {
        let network = context.config.network();
        let sha = sha2::Sha256::digest(&pubkey);
        let pkh: [u8; 20] = ripemd::Ripemd160::digest(&sha).into();
        let tkey = TransparentAddress::PublicKeyHash(pkh);
        let taddr = zcash_client_backend::address::Address::Transparent(tkey);
        let taddr = taddr.encode(&network);
        taddr
    });
    Ok(taddr)
}

pub fn get_ovk(pubkey: Vec<u8>) -> Result<Vec<u8>, ZcashError> {
    let hash = blake2b_simd::Params::new()
        .hash_length(32)
        .personal(b"Zcash_Maya_OVK__")
        .hash(&pubkey);
    let ovk = hash.as_bytes().to_vec();
    Ok(ovk)
}

pub fn validate_address(address: String) -> Result<bool, ZcashError> {
    uniffi_export!(context, {
        let network = context.config.network();
        let r = Address::decode(&network, &address);
        let res = match r {
            // TEX addresses are for centralized exchanges (i.e.
            // Binance); there is no reason to support them.
            Some(Address::Tex(_)) => false,
            Some(_) => true,
            None => false,
        };
        Ok(res)
    })
}

pub fn match_with_blockchain_receiver(
    address: String,
    receiver: String,
) -> Result<bool, ZcashError> {
    uniffi_export!(context, {
        let network = context.config.network();
        let address_receivers = extract_receivers(&network, &address)?;
        let receivers = extract_receivers(&network, &receiver)?;
        if receivers.len() != 1 {
            return Err(ZcashError::AssertError(
                "Blockchain address must have a single receiver".to_string(),
            ));
        }
        let contains = address_receivers.contains(receivers.first().unwrap());

        Ok::<_, ZcashError>(contains)
    })
}

fn extract_receivers(network: &Network, address: &str) -> Result<Vec<Receiver>, ZcashError> {
    let receiver = Address::decode(network, address)
        .ok_or_else(|| ZcashError::InvalidAddress(address.to_string()))?;
    let receivers = match receiver {
        Address::Transparent(transparent_address) => match transparent_address {
            TransparentAddress::PublicKeyHash(pkh) => vec![unified::Receiver::P2pkh(pkh)],
            TransparentAddress::ScriptHash(sh) => vec![unified::Receiver::P2sh(sh)],
        },
        Address::Sapling(payment_address) => {
            vec![unified::Receiver::Sapling(payment_address.to_bytes())]
        }
        Address::Unified(_) => {
            let (_, ua) = unified::Address::decode(address).unwrap();
            ua.items()
        }
        Address::Tex(_) => {
            return Err(ZcashError::InvalidAddress(address.to_string()));
        }
    };
    Ok(receivers)
}
