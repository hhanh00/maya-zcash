use anyhow::Result;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::{config::Config, ZcashError};

#[derive(Serialize)]
pub struct RpcRequest<'a> {
    jsonrpc: &'a str,
    id: &'a str,
    method: &'a str,
    params: Vec<Value>,
}

// JSON-RPC response structure
#[derive(Deserialize, Debug)]
pub struct RpcResponse {
    result: serde_json::Value,
    error: Option<serde_json::Value>,
}

pub async fn json_request<'a>(
    config: &Config,
    id: &'a str,
    method: &'a str,
    params: Vec<Value>,
) -> Result<Value> {
    // Create the JSON-RPC payload
    let req = RpcRequest {
        jsonrpc: "1.0",
        id,
        method,
        params,
    };
    // Create an HTTP client
    let client = Client::new();
    // Send the request
    let response = client
        .post(&config.server.host)
        .basic_auth(&config.server.user, Some(&config.server.password)) // Add authentication
        .json(&req) // Send JSON payload
        .send()
        .await?;

    // Deserialize the response
    let rpc_response: RpcResponse = response.json().await?;
    // Print the response
    if let Some(error) = rpc_response.error {
        tracing::error!("Error: {:?}", error);
        anyhow::bail!(error["message"].as_str().unwrap().to_string());
    }
    Ok(rpc_response.result)
}

pub fn map_rpc_error(e: anyhow::Error) -> ZcashError {
    ZcashError::RPC(e.to_string())
}
