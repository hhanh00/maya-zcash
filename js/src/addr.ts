import { secp256k1 } from "@noble/curves/secp256k1";
import { hmac } from '@noble/hashes/hmac';
import { sha256 } from '@noble/hashes/sha2';
import { ripemd160 } from '@noble/hashes/ripemd160';
import bs58check from 'bs58check';

export function skToAddr(sk: Uint8Array, prefix: Uint8Array) {
    const pk = secp256k1.getPublicKey(sk, true);

    const hash = sha256(pk);
    const pkh = ripemd160(hash)

    const addrb = Buffer.alloc(22);
    Buffer.from(prefix).copy(addrb);
    Buffer.from(pkh).copy(addrb, 2);
    const addr = bs58check.encode(addrb);
}
