pub mod config;
pub mod rpc;
pub mod chain;
pub mod addr;

use std::path::Path;

use config::Context;
use parking_lot::Mutex;
use thiserror::Error;
use tokio::runtime::Runtime;

#[derive(Debug, Error)]
pub enum ZcashError {
    #[error("RPC Error: {0}")]
    RPC(String),
    #[error("PubKey must be 33 bytes long: {0}")]
    InvalidPubkeyLength(String),
}

pub struct Height {
    number: u32,
    hash: Vec<u8>,
}

lazy_static::lazy_static! {
    pub static ref CONTEXT: Mutex<Context> = Mutex::new(init());
}

fn init() -> Context {
    let config = config::read_config(Path::new("config.yaml")).expect("Missing config.yaml");
    let runtime = Runtime::new().unwrap();
    let context = Context {
        config,
        runtime,
    };
    context
}

use crate::chain::get_latest_height;
use crate::addr::{get_vault_address, validate_address};

uniffi::include_scaffolding!("interface");

#[macro_export]
macro_rules! uniffi_export {
    ($config:ident, $block:block) => {
        {
            let context = crate::CONTEXT.lock();
            let $config = &context.config;
            $block
        }
    };
}

#[macro_export]
macro_rules! uniffi_async_export {
    ($config:ident, $block:block) => {
        {
            let context = crate::CONTEXT.lock();
            let $config = &context.config;
            context.runtime.block_on(async {
                let res = $block;
                Ok::<_, ZcashError>(res)
            })
        }
    };
}
