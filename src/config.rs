use anyhow::Result;
use figment::{providers::{Format as _, Yaml}, Figment};
use tokio::runtime::Runtime;
use std::{fs::File, io::Read, path::Path};

use serde::Deserialize;

#[derive(Deserialize, Debug)]
pub struct Server {
    pub host: String,
    pub user: String,
    pub password: String,
}

#[derive(Deserialize, Debug)]
pub struct Config {
    pub server: Server,
}

pub struct Context {
    pub config: Config,
    pub runtime: Runtime,
}

pub fn read_config(path: &Path) -> Result<Config> {
    let mut p = File::open(path)?;
    let mut buf = String::new();
    p.read_to_string(&mut buf)?;
    let config = Figment::new()
    .merge(Yaml::string(&buf))
    .extract::<Config>()?;

    Ok(config)
}
