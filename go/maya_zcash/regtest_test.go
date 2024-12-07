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
