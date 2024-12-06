use sha2::Digest as _;
use zcash_address::unified::{self, Container, Encoding as _, Receiver};
use zcash_keys::address::Address;
use zcash_primitives::legacy::TransparentAddress;
use zcash_protocol::consensus::Network;

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

pub fn match_with_blockchain_receiver(address: String, receiver: String) -> Result<bool, ZcashError> {
    uniffi_export!(config, {
        let network = config.network();
        let address_receivers = extract_receivers(&network, &address)?;
        let receivers = extract_receivers(&network, &receiver)?;
        if receivers.len() != 1 {
            return Err(ZcashError::AssertError("Blockchain address must have a single receiver".to_string()));
        }
        let contains = address_receivers.contains(receivers.first().unwrap());

        Ok::<_, ZcashError>(contains)
    })
}

fn extract_receivers(network: &Network, address: &str) -> Result<Vec<Receiver>, ZcashError> {
    let receiver = Address::decode(network, address).ok_or_else(|| ZcashError::InvalidAddress(address.to_string()))?;
    let receivers = match receiver {
        Address::Transparent(transparent_address) => {
            match transparent_address {
                TransparentAddress::PublicKeyHash(pkh) => vec![unified::Receiver::P2pkh(pkh)],
                TransparentAddress::ScriptHash(sh) => vec![unified::Receiver::P2sh(sh)],
            }
        }
        Address::Sapling(payment_address) => {
            vec![unified::Receiver::Sapling(payment_address.to_bytes())]
        }
        Address::Unified(_) => {
            let (_, ua) = unified::Address::decode(address).unwrap();
            ua.items()
        },
        Address::Tex(_) => {
            return Err(ZcashError::InvalidAddress(address.to_string()));
        },
    };
    Ok(receivers)
}
