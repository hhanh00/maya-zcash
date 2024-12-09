package maya_zcash

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestMain(t *testing.M) {
    InitLogger();
    t.Run();
}

func TestLatestHeight(t *testing.T) {
    height, err := GetLatestHeight()
    if height.Number < 100 {
        t.Errorf(`GetLatestHeight = %v, %v`, height, err)
    }
}

func TestVaultAddress(t *testing.T) {
    bytes, _ := hex.DecodeString("03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7")
    address, err := GetVaultAddress(bytes)
    if err != nil {
        t.Errorf(`TestVaultAddress = %v`, err)
    }
    fmt.Printf("address: %v\n", address)
    // regtest: tmGys6dBuEGjch5LFnhdo5gpSa7jiNRWse6
}

func TestValidateAddress(t *testing.T) {
    valid, err := ValidateAddress("tmWksakBYGg7Lqtm1EqSqvPkVYJHYxGq6Za")
    if err != nil {
        t.Errorf(`TestValidateAddress = %v`, err)
    }
    if !valid {
        t.Errorf("Should be a valid address")
    }

    valid, err = ValidateAddress("t1invalidaddress")
    if err != nil {
        t.Errorf(`TestValidateAddress = %v`, err)
    }
    if valid {
        t.Errorf("Should be a invalid address")
    }
}

func TestMatchWithBlockchainReceiver(t *testing.T) {
    valid, err := MatchWithBlockchainReceiver("tmWksakBYGg7Lqtm1EqSqvPkVYJHYxGq6Za", "tmWksakBYGg7Lqtm1EqSqvPkVYJHYxGq6Za")
    if err != nil {
        t.Errorf(`TestMatchWithBlockchainReceiver = %v`, err)
    }
    if !valid {
        t.Errorf("Address should contain itself")
    }

    // Add test UA contains Sap
    // Add test UA not contains Orc
}

func TestBalance(t *testing.T) {
    balance, err := GetBalance("tmGys6dBuEGjch5LFnhdo5gpSa7jiNRWse6")
    if err != nil {
        t.Errorf(`TestBalance = %v`, err)
    }
    if balance == 0 {
        t.Errorf("This address should have some funds")
    }
}

func TestListUTXO(t *testing.T) {
    utxos, err := ListUtxos("tmGys6dBuEGjch5LFnhdo5gpSa7jiNRWse6")
    if err != nil {
        t.Errorf(`TestListUTXO = %v`, err)
    }
    if len(utxos) == 0 {
        t.Errorf("Must have at least one UTXO")
    }
    if utxos[0].Value != 540000000 {
        t.Errorf("UTXO value does not match")
    }
}

func TestScanMempool(t *testing.T) {
    bytes, _ := hex.DecodeString("03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7")
    _, err := ScanMempool(bytes)
    if err != nil {
        t.Errorf(`TestScanMempool = %v`, err)
    }
}

func TestSKToPub(t *testing.T) {
    // This is the secret key of the vault
    k, err := SkToPub("L1rrP7J2tqVfC5sj5wi8Gn4M2f4kyX1dByHPVHCa6Mzyz8eahu77")
    if err != nil {
        t.Errorf(`TestSKToPub = %v`, err)
    }
    sk := hex.EncodeToString(k.Sk)
    pk := hex.EncodeToString(k.Pk)
    fmt.Printf("sk: %s pk: %v addr %s\n", sk, pk, k.Addr)
}

func TestSendToVault(t *testing.T) {
    // secret key of the user account: L1sjrupHTXwtX847jZhXpkACVYE6d4edPeJK9762j7AeCYL4c32z
    sk, _ := hex.DecodeString("8ae9c0c958937eeec71e034650e889085c10e91ae1ab94a26c26182f9516a37f")
    vault, _ := hex.DecodeString("03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7")
    _, err := SendToVault(200, sk, "tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU", vault, 10000000, "MEMO")
    if err != nil {
        t.Errorf(`TestSendToVault = %v`, err)
    }
}

func TestBroadcast(t *testing.T) {
    // Skip this test because hardcoding a raw tx does not work
    t.SkipNow()

    txb, _ := hex.DecodeString("050000800a27a7265510e7c800000000f00000000186b1a9c7f46c7550e48fa0781495ef891ed81b48506ef53a28c3e83a223f6482000000006a473044022014e4bc7f7ab7034fe1992ee128484e72864e7d08380b2e663b43c2d490aa26190220643a69a289195a62f9c85c208a27d417f847fd1ea94af086cce435ce6eb9f89f012103243597856d5bd7c8f91f77446a53db425ce10d237c1d6928f2268acdc538797effffffff030000000000000000066a044d454d4f80969800000000001976a914e6d4b9d2c408bf6bd44523b3b6607de4853b806088ace83c8e06000000001976a914936667ff8d2d41361a4df4a370b309fb15380eac88ac0000002024")
    txid, err := BroadcastRawTx(txb)
    if err != nil {
        t.Errorf(`TestSendToVault = %v`, err)
    }
    if txid != "d120e67dac6ccdb49915542544ae2673fd4aef6adc1fa4eac9012134c9f3ddd0" {
        t.Errorf(`txid mismatch = %s`, txid)
    }
}

func TestPayFromVault(t *testing.T) {
    vault, _ := hex.DecodeString("03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7")
    ptx, err := PayFromVault(200, vault, "tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU", 500000, "MEMO OUT")
    if err != nil {
        t.Errorf(`TestPayFromVault = %v`, err)
    }
    if ptx.Fee != 15000 {
        t.Errorf(`Unexpected fee %d`, ptx.Fee)
    }
}

func TestCombineVault(t *testing.T) {
    vault, _ := hex.DecodeString("03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7")
    ptx, err := CombineVault(200, vault)
    if err != nil {
        t.Errorf(`TestCombineVault = %v`, err)
    }
    if ptx.Outputs[0].Amount != 539995000 {
        t.Errorf(`Unexpected amount %d`, ptx.Outputs[0].Amount)
    }
}

func TestBuildVaultUnauthorizedTx(t *testing.T) {
    vault, _ := hex.DecodeString("03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7")
    ptx, _ := PayFromVault(200, vault, "zregtestsapling18ywlqhk60zglax5drk3kwltkmcatf5eptxyrkrx20hcqma5nsvrgh63843seye923qk5wfvxpnr", 500000, "MEMO OUT")
    sighashes, err := BuildVaultUnauthorizedTx(vault, ptx)
    if err != nil {
        t.Errorf(`TestBuildVaultUnauthorizedTx = %v`, err)
    }
    fmt.Printf("txseed: %v\n", hex.EncodeToString(ptx.TxSeed))
    fmt.Printf("sighash[0]: %v\n", hex.EncodeToString(sighashes.Hashes[0]))
}

func TestSignSighash(t *testing.T) {
    sk, _ := hex.DecodeString("8ae9c0c958937eeec71e034650e889085c10e91ae1ab94a26c26182f9516a37f")
    sighash, _ := hex.DecodeString("32fe38e61df5290198ec736e7b0a1b7cb8a372e42d26c2e3aabcfed29977e911")
    signature, err := SignSighash(sk, sighash)
    if err != nil {
        t.Errorf(`TestSignSighash = %v`, err)
    }
    fmt.Printf("signature: %v\n", hex.EncodeToString(signature))
}

func TestApplySignatures(t *testing.T) {
    vault_sk, _ := hex.DecodeString("8a74dce839bc2228428ed5de3c2edbabb5c9713f5e6eeb808f9c56640921c6c9")
    vault, _ := hex.DecodeString("03c622fa3be76cd25180d5a61387362181caca77242023be11775134fd37f403f7")
    ptx, _ := PayFromVault(200, vault, "zregtestsapling18ywlqhk60zglax5drk3kwltkmcatf5eptxyrkrx20hcqma5nsvrgh63843seye923qk5wfvxpnr", 500000, "MEMO OUT")
    sighashes, _ := BuildVaultUnauthorizedTx(vault, ptx)
    signatures := make([][]byte, 0)
    for _, sighash := range sighashes.Hashes {
        signature, _ := SignSighash(vault_sk, sighash)
        signatures = append(signatures, signature)
    }
    txb, err := ApplySignatures(vault, ptx, signatures)
    if err != nil {
        t.Errorf(`TestApplySignatures = %v`, err)
    }
    // fmt.Printf("txb: %v\n", hex.EncodeToString(txb))
    txid, _ := BroadcastRawTx(txb)
    fmt.Printf("txid: %s\n", txid)
}
