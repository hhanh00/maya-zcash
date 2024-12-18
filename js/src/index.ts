
import blake2b from 'blake2b-wasm';
import { Config } from './types';
import { buildTx, signAndFinalize } from './builder';
import { sendRawTransaction } from './rpc';

const config: Config = {
    server: {
        host: "http://172.16.11.111:18232",
        user: "mayachain",
        password: "password"
    },
    mainnet: false
}

const address = 'tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU';

async function main() {
    await blake2b.ready();
    const utx = await buildTx(200, 'tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU', 'tmGys6dBuEGjch5LFnhdo5gpSa7jiNRWse6', 1000000,
        'MEMO', config
    )
    const txb = await signAndFinalize(utx.height, '8ae9c0c958937eeec71e034650e889085c10e91ae1ab94a26c26182f9516a37f',
        utx.inputs, utx.outputs);
    const txid = await sendRawTransaction(txb, config);
    console.log(txid);
}

main()

