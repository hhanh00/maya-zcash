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
    bytes, _ := hex.DecodeString("02c72d6f1a74d169ddbdf5b7da258ece5fa09cc6b13385a8b0bcd7b1aef3bf4483")
    address, err := GetVaultAddress(bytes)
    if err != nil {
        t.Errorf(`TestVaultAddress = %v`, err)
    }
    fmt.Printf("address: %v\n", address)
    // regtest: tmWksakBYGg7Lqtm1EqSqvPkVYJHYxGq6Za
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
    balance, err := GetBalance("tmWksakBYGg7Lqtm1EqSqvPkVYJHYxGq6Za")
    if err != nil {
        t.Errorf(`TestBalance = %v`, err)
    }
    if balance == 0 {
        t.Errorf("This address should have some funds")
    }
}

func TestListUTXO(t *testing.T) {
    utxos, err := ListUtxos("tmWksakBYGg7Lqtm1EqSqvPkVYJHYxGq6Za")
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
    bytes, _ := hex.DecodeString("02c72d6f1a74d169ddbdf5b7da258ece5fa09cc6b13385a8b0bcd7b1aef3bf4483")
    _, err := ScanMempool(bytes)
    if err != nil {
        t.Errorf(`TestScanMempool = %v`, err)
    }
}

func TestSKToPub(t *testing.T) {
    k, err := SkToPub("L1sjrupHTXwtX847jZhXpkACVYE6d4edPeJK9762j7AeCYL4c32z")
    if err != nil {
        t.Errorf(`TestSKToPub = %v`, err)
    }
    sk := hex.EncodeToString(k.Sk)
    fmt.Printf("sk: %s k: %v\n", sk, k)
}

func TestSendToVault(t *testing.T) {
    sk, _ := hex.DecodeString("8ae9c0c958937eeec71e034650e889085c10e91ae1ab94a26c26182f9516a37f")
    vault, _ := hex.DecodeString("02c72d6f1a74d169ddbdf5b7da258ece5fa09cc6b13385a8b0bcd7b1aef3bf4483")
    tx, err := SendToVault(200, sk, "tmP9jLgTnhDdKdWJCm4BT2t6acGnxqP14yU", vault, 10000000, "MEMO")
    if err != nil {
        t.Errorf(`TestSendToVault = %v`, err)
    }
    txb := tx.Data
    txid, err := BroadcastRawTx(txb)
    if err != nil {
        t.Errorf(`TestSendToVault = %v`, err)
    }
    fmt.Printf("txid: %v\n", txid)
}
