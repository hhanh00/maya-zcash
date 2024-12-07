use crate::ZcashError;

pub struct Note {
    pub address: String,
    pub value: i64,
    pub memo: Option<String>,
}

pub struct TxData {
    pub txid: String,
    pub height: u32, // block height or 0 if unconfirmed
    // Who paid us or who we are paying, and how much
    // value > 0 -> vault received funds
    pub counterparty: Note,
    // All the transparent input/output and the shielded outputs
    // we can decrypt
    pub plain: Vec<Note>,
    // amount encrypted
    pub encrypted: i64,
    pub fee: u64,
}

pub fn scan_mempool(pubkey: Vec<u8>) -> Result<Vec<TxData>, ZcashError> {
    // pubkey -> taddr, ovk
    // for each tx
    // - decode into t/z/o bundles
    // - check tbundle
    //   - i/o not contains taddr -> continue
    //   - resolve inp (txid, vout) -> (addr, value)
    //   - decode OP_RESULT if any
    // - check z/o bundles
    //   - try decrypt out with ovk
    //   - if tinp contains taddr, every output should be decryptable
    //     (addr, value)
    //   - inp are always unknown

    Ok(vec![])
}
