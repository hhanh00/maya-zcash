export type Config = {
    server: {
        host: string;
        user: string;
        password: string;
    };
    mainnet: boolean;
}

type UTXO = {
    address: string;
    txid: string;
    outputIndex: number;
    satoshis: number;
}

type OutputPKH = {
    type: 'pkh';
    address: string;
    amount: number;
}

type OutputMemo = {
    type: 'op_return';
    memo: string;
}

type Output = OutputPKH | OutputMemo;

type Tx = {
    height: number;
    inputs: UTXO[];
    outputs: Output[];
}
