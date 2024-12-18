
import blake2b from 'blake2b-wasm';
import { Config, pkToAddr, testnetPrefix, sendRawTransaction, 
    buildTx, signAndFinalize } from '.';

const config: Config = {
    server: {
        host: "http://172.16.11.111:18232",
        user: "mayachain",
        password: "password"
    },
    mainnet: false
}

const address = 'tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU';

test('get address of vault', () => {
    const addr = pkToAddr(
        Buffer.from('03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7', 'hex'),
        Buffer.from(testnetPrefix));
    expect(addr).toBe('tmGys6dBuEGjch5LFnhdo5gpSa7jiNRWse6');
});

test('tx fee', async () => {
    const utx = await buildTx(200, 'tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU', 'tmGys6dBuEGjch5LFnhdo5gpSa7jiNRWse6', 1000000,
        'MEMO', config)
    expect(utx.fee).toBe(15000)
});

test('create/send t2t tx', async () => {
    await blake2b.ready();
    const utx = await buildTx(200, 'tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU', 'tmGys6dBuEGjch5LFnhdo5gpSa7jiNRWse6', 1000000,
        'MEMO', config)
    const txb = await signAndFinalize(utx.height, '8ae9c0c958937eeec71e034650e889085c10e91ae1ab94a26c26182f9516a37f',
        utx.inputs, utx.outputs);
    const txid = await sendRawTransaction(txb, config);
    expect(typeof txid).toBe("string")
    console.log(txid);
})
