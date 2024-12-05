pub mod config;
pub mod rpc;
pub mod chain;

use std::path::Path;

use config::Context;
use parking_lot::Mutex;
use thiserror::Error;
use tokio::runtime::Runtime;

#[derive(Debug, Error)]
pub enum ZcashError {
    #[error("RPC Error: {0}")]
    RPC(String),
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

uniffi::include_scaffolding!("interface");

#[macro_export]
macro_rules! uniffi_export {
    ($config:ident, $block:block) => {
        {
            let context = CONTEXT.lock();
            let $config = &context.config;
            context.runtime.block_on(async {
                let res = $block;
                Ok::<_, ZcashError>(res)
            })
        }
    };
}
