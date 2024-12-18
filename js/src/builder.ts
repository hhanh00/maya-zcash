import blake2b from 'blake2b-wasm';
import { max, min, sumBy } from "lodash";
import { secp256k1 } from '@noble/curves/secp256k1';
import { addressToScript, memoToScript, writeSigScript } from "./script";
import { writeCompactInt } from "./writer";
import { Config, Output, UTXO } from "./types";
import { getUTXOS } from "./rpc";
import { isValidAddr, mainnetPrefix, testnetPrefix } from './addr';

function getFee(utxos: UTXO[]): number {
    const BASE_FEE = 5000;
    return max([utxos.length, 3])! * BASE_FEE;
}

function selectUTXOS(utxos: UTXO[], amount: number): UTXO[] {
    var currentFee = 0
    var selected = []
    var remaining = amount;

    for (const utxo of utxos) {
        if (remaining == 0) break;

        selected.push(utxo);
        const fee = getFee(selected);
        const deltaFee = fee - currentFee;
        currentFee = fee;
        remaining += deltaFee;

        const used = min([utxo.satoshis, remaining])!;
        remaining -= used;
    }

    return selected;
}

// @ts-ignore
export async function buildTx(height: number, from: string, to: string, amount: number, memo: string, config: Config): Promise<Tx> {
    const prefixb = config.mainnet ? mainnetPrefix : testnetPrefix;
    const prefix = Buffer.from(prefixb);
    if (!isValidAddr(from, prefix)) throw new Error('Invalid "from" address');
    if (!isValidAddr(to, prefix)) throw new Error('Invalid "to" address');
    if (amount > 1e14) throw new Error('Amount too large');
    if (memo.length > 80) throw new Error('Memo too long');

    const utxos = await getUTXOS(from, config);
    const inputs = selectUTXOS(utxos, amount);
    const fee = getFee(utxos);
    const change = sumBy(inputs, (u) => u.satoshis) - amount - fee;
    if (change < 0)
        throw new Error('Not enough funds');

    var outputs: Output[] = [];
    outputs.push({
        type: 'pkh',
        address: from,
        amount: change,
    });
    outputs.push({
        type: 'pkh',
        address: to,
        amount: amount,
    });
    outputs.push({
        type: 'op_return',
        memo: memo,
    });
    return {
        height: height,
        inputs: inputs,
        outputs: outputs,
        fee: fee,
    }
}

export async function signAndFinalize(height: number, skb: string, utxos: UTXO[], outputs: Output[]): Promise<Buffer> {
    const sk = new Uint8Array(Buffer.from(skb, 'hex'));
    const pk = secp256k1.getPublicKey(sk, true);

    var offset = 0;

    // HEADER
    var buf = Buffer.alloc(20);
    buf.writeUInt32LE(0x80000005, 0);
    buf.writeUInt32LE(0x26a7270a, 4);
    buf.writeUInt32LE(0xc8e71055, 8);
    buf.writeUInt32LE(0x00000000, 12);
    buf.writeUInt32LE(height, 16);

    var h = blake2b(32, undefined, undefined,
        new TextEncoder().encode('ZTxIdHeadersHash')
    )
    h.update(buf);
    const headerHash = h.digest('hex')

    buf = Buffer.alloc(36 * utxos.length);
    for (const [i, utxo] of utxos.entries()) {
        const txid = Buffer.from(utxo.txid, 'hex');
        txid.reverse();
        txid.copy(buf, 36 * i);
        buf.writeUInt32LE(utxo.outputIndex, 36 * i + 32);
    }
    h = blake2b(32, undefined, undefined,
        new TextEncoder().encode('ZTxIdPrevoutHash')
    )
    h.update(buf);
    const prevoutputsHash = h.digest('hex')

    buf = Buffer.alloc(4 * utxos.length);
    for (const [i, _] of utxos.entries()) {
        buf.writeInt32LE(-1, 4 * i);
    }
    h = blake2b(32, undefined, undefined,
        new TextEncoder().encode('ZTxIdSequencHash')
    )
    h.update(buf);
    const sequencesHash = h.digest('hex')

    buf = Buffer.alloc(34 * outputs.length);
    offset = 0;
    for (const [i, output] of outputs.entries()) {
        switch (output.type) {
            case 'pkh':
                buf.writeUIntLE(output.amount, offset, 6); // 6 is the max
                offset += 8;
                const pkhscript = addressToScript(output.address);
                pkhscript.copy(buf, offset);
                offset += 26;
                break;

            case "op_return":
                offset += 8;
                const oprscript = memoToScript(output.memo);
                oprscript.copy(buf, offset);
                offset += oprscript.length;
                break;
        }
    }
    h = blake2b(32, undefined, undefined,
        new TextEncoder().encode('ZTxIdOutputsHash')
    )
    h.update(buf.subarray(0, offset));
    const outputsHash = h.digest('hex')

    buf = Buffer.alloc(8 * utxos.length);
    for (const [i, utxo] of utxos.entries()) {
        buf.writeUIntLE(utxo.satoshis, 8 * i, 6);
    }
    h = blake2b(32, undefined, undefined,
        new TextEncoder().encode('ZTxTrAmountsHash')
    )
    h.update(buf);
    const amountsHash = h.digest('hex')

    buf = Buffer.alloc(26 * utxos.length);
    for (const [i, utxo] of utxos.entries()) {
        const script = addressToScript(utxo.address);
        script.copy(buf, 26 * i);
    }
    h = blake2b(32, undefined, undefined,
        new TextEncoder().encode('ZTxTrScriptsHash')
    )
    h.update(buf);
    const scriptsHash = h.digest('hex')

    const signatures: Uint8Array[] = [];
    for (const [i, utxo] of utxos.entries()) {
        buf = Buffer.alloc(32 + 4 + 8 + 26 + 4);
        offset = 0;
        const txid = Buffer.from(utxo.txid, 'hex');
        txid.reverse();
        txid.copy(buf, offset);
        offset += 32;
        buf.writeUInt32LE(utxo.outputIndex, offset);
        offset += 4;
        buf.writeUIntLE(utxo.satoshis, offset, 6);
        offset += 8;
        const script = addressToScript(utxo.address);
        script.copy(buf, offset);
        offset += 26;
        buf.writeInt32LE(-1, offset);

        h = blake2b(32, undefined, undefined,
            new TextEncoder().encode('Zcash___TxInHash')
        )
        h.update(buf);
        const txInHash = h.digest('hex')

        buf = Buffer.alloc(1 + 32 * 6);
        offset = 1;
        buf[0] = 1;
        Buffer.from(prevoutputsHash, 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from(amountsHash, 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from(scriptsHash, 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from(sequencesHash, 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from(outputsHash, 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from(txInHash, 'hex').copy(buf, offset);
        offset += 32;

        h = blake2b(32, undefined, undefined,
            new TextEncoder().encode('ZTxIdTranspaHash')
        )
        h.update(buf);
        const transparentHash = h.digest('hex')

        buf = Buffer.alloc(32 * 4);
        offset = 0;
        Buffer.from(headerHash, 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from(transparentHash, 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from('6f2fc8f98feafd94e74a0df4bed74391ee0b5a69945e4ced8ca8a095206f00ae', 'hex').copy(buf, offset);
        offset += 32;
        Buffer.from('9fbe4ed13b0c08e671c11a3407d84e1117cd45028a2eee1b9feae78b48a6e2c1', 'hex').copy(buf, offset);
        offset += 32;

        const personal = Buffer.alloc(16);
        Buffer.from('ZcashTxHash_').copy(personal);
        personal.writeUInt32LE(0xc8e71055, 12);

        h = blake2b(32, undefined, undefined,
            personal
        )
        h.update(buf);
        const sigHash = h.digest()

        const signature = await secp256k1.sign(sigHash, sk, { lowS: true, prehash: false });
        const signatureDER = signature.toDERRawBytes()
        signatures.push(signatureDER);
    }

    buf = Buffer.alloc(2000);
    offset = 0;

    buf.writeUInt32LE(0x80000005, offset);
    offset += 4;
    buf.writeUInt32LE(0x26a7270a, offset);
    offset += 4;
    buf.writeUInt32LE(0xc8e71055, offset);
    offset += 4;
    buf.writeUInt32LE(0x00000000, offset);
    offset += 4;
    buf.writeUInt32LE(height, offset);
    offset += 4;

    const txinc = writeCompactInt(utxos.length);
    txinc.copy(buf, offset);
    offset += txinc.length;
    for (const [i, utxo] of utxos.entries()) {
        const txid = Buffer.from(utxo.txid, 'hex');
        txid.reverse();
        txid.copy(buf, offset);
        offset += 32;
        buf.writeUInt32LE(utxo.outputIndex, offset);
        offset += 4;
        const ss = writeSigScript(signatures[i], pk);
        const ssl = writeCompactInt(ss.length);
        ssl.copy(buf, offset);
        offset += ssl.length;
        ss.copy(buf, offset);
        offset += ss.length;
        buf.writeInt32LE(-1, offset);
        offset += 4;
    }

    const txoutc = writeCompactInt(outputs.length);
    txoutc.copy(buf, offset);
    offset += txoutc.length;
    for (const [i, out] of outputs.entries()) {
        switch (out.type) {
            case "pkh": {
                buf.writeBigInt64LE(BigInt(out.amount), offset);
                offset += 8;
                const pkhscript = addressToScript(out.address);
                pkhscript.copy(buf, offset);
                offset += pkhscript.length;
            }
                break;
            case "op_return": {
                offset += 8;
                const memoscript = memoToScript(out.memo);
                memoscript.copy(buf, offset);
                offset += memoscript.length;
            }
                break;
        }
    }
    // Add 000000
    offset += 3;

    return buf.subarray(0, offset);
}
