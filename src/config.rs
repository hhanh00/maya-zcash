use std::{collections::HashMap, env, fs::File, io::Read};

use anyhow::Result;
use tokio::runtime::Runtime;

use serde::Deserialize;

use crate::network::Network;

#[derive(Deserialize, Debug)]
pub struct Server {
    pub host: String,
    pub user: String,
    pub password: String,
}

#[derive(Deserialize, Debug)]
pub struct Config {
    pub server: Server,
    pub mainnet: bool,
    pub sapling_params_dir: String,
}

impl Config {
    pub fn network(&self) -> Network {
        if self.mainnet {
            Network::Main
        } else {
            Network::Regtest
        }
    }
}

pub struct Context {
    pub config: Config,
    pub runtime: Runtime,
}

pub fn read_config(name: &str) -> Result<Config> {
    let mut p = File::open(name)?;
    let mut s = String::new();
    p.read_to_string(&mut s)?;

    let env_vars: HashMap<String, String> = env::vars().collect();
    let config = subst::yaml::from_str::<Config, _>(&s, &env_vars)?;

    Ok(config)
}
