import { secp256k1 } from "@noble/curves/secp256k1";
import { hmac } from '@noble/hashes/hmac';
import { sha256 } from '@noble/hashes/sha2';
import { ripemd160 } from '@noble/hashes/ripemd160';
import bs58check from 'bs58check';

export const testnetPrefix = [29, 37];
export const mainnetPrefix = [28, 189];

export function skToAddr(sk: Uint8Array, prefix: Uint8Array) {
    const pk = secp256k1.getPublicKey(sk, true);
    return pkToAddr(pk, prefix);
}

export function pkToAddr(pk: Uint8Array, prefix: Uint8Array) {
    const hash = sha256(pk);
    const pkh = ripemd160(hash)

    const addrb = Buffer.alloc(22);
    Buffer.from(prefix).copy(addrb);
    Buffer.from(pkh).copy(addrb, 2);
    const addr = bs58check.encode(addrb);

    return addr;
}
