import { JSONRPCClient } from "json-rpc-2.0";
import { Config, UTXO } from "./types";
import axios from "axios";

function makeClient(config: Config): JSONRPCClient {
    const client = new JSONRPCClient(async (jsonRPCRequest) => {
        const response = await axios.post(config.server.host, jsonRPCRequest, {
            headers: { "Content-Type": "application/json" },
            auth: {
                username: config.server.user,
                password: config.server.password,
            },
        });
        client.receive(response.data);
    });

    return client;
}

export async function getUTXOS(from: string, config: Config): Promise<UTXO[]> {
    const client = makeClient(config);
    const utxos: UTXO[] = await client.request('getaddressutxos', [from]);
    return utxos;
}

export async function sendRawTransaction(txb: Buffer, config: Config) {
    const client = makeClient(config);
    const txid = await client.request('sendrawtransaction', [txb.toString('hex')]);
    return txid
}
