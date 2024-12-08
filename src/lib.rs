pub mod addr;
pub mod chain;
pub mod config;
pub mod network;
pub mod pay;
pub mod rpc;
pub mod scan;
pub mod wallet;

use config::Context;
use parking_lot::ReentrantMutex;
use thiserror::Error;
use tokio::runtime::Runtime;
use tracing_subscriber::layer::SubscriberExt as _;
use tracing_subscriber::util::SubscriberInitExt as _;
use tracing_subscriber::{fmt, EnvFilter};

#[derive(Debug, Error)]
pub enum ZcashError {
    #[error("RPC Error: {0}")]
    RPC(String),
    #[error("PubKey must be 33 bytes long: {0}")]
    InvalidPubkeyLength(String),
    #[error("Invalid address: {0}")]
    InvalidAddress(String),
    #[error("No Orchard receiver")]
    NoOrchardReceiver,
    #[error("No enough funds")]
    NotEnoughFunds,
    #[error("Transaction rejected by server")]
    TxRejected,
    #[error("Assertion Failed: {0}")]
    AssertError(String),
}

pub struct Height {
    number: u32,
    hash: Vec<u8>,
}

lazy_static::lazy_static! {
    pub static ref CONTEXT: ReentrantMutex<Context> = ReentrantMutex::new(init());
}

fn init() -> Context {
    let config = config::read_config("config.yaml").expect("Missing config.yaml");
    let runtime = Runtime::new().unwrap();
    let context = Context { config, runtime };
    context
}

pub fn init_logger() {
    tracing_subscriber::registry()
        .with(fmt::layer())
        .with(EnvFilter::from_default_env())
        .init();
}

use crate::addr::{get_ovk, get_vault_address, match_with_blockchain_receiver, validate_address};
use crate::chain::{broadcast_raw_tx, get_latest_height};
use crate::pay::{send_to_vault, TxBytes};
use crate::scan::{scan_mempool, Note, TxData};
use crate::wallet::{get_balance, list_utxos, sk_to_pub, TransparentKey, UTXO};

uniffi::include_scaffolding!("interface");

#[macro_export]
macro_rules! uniffi_export {
    ($config:ident, $block:block) => {{
        let context = crate::CONTEXT.lock();
        let $config = &context.config;
        $block
    }};
}

#[macro_export]
macro_rules! uniffi_async_export {
    ($config:ident, $block:block) => {{
        let context = crate::CONTEXT.lock();
        let $config = &context.config;
        context.runtime.block_on(async { $block })
    }};
}

pub fn decode_hexstring(s: &str) -> Result<Vec<u8>, ZcashError> {
    hex::decode(s).map_err(|_| ZcashError::AssertError("Invalid Hex string".into()))
}

pub fn to_ba<const N: usize>(v: Vec<u8>) -> Result<[u8; N], ZcashError> {
    let v: Result<[u8; N], _> = v.try_into();
    v.map_err(|_| ZcashError::AssertError("".into()))
}

pub fn to_hash(s: &str) -> Result<[u8; 32], ZcashError> {
    let mut v = decode_hexstring(s)?;
    v.reverse();
    to_ba(v)
}

pub fn to_zcasherror<E>(anyerror: anyhow::Error) -> impl FnOnce(E) -> ZcashError {
    let map = move |_| ZcashError::AssertError(anyerror.to_string());
    map
}

impl From<anyhow::Error> for ZcashError {
    fn from(value: anyhow::Error) -> Self {
        ZcashError::AssertError(value.to_string())
    }
}
