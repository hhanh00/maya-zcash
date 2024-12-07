package maya_zcash

import (
	"encoding/hex"
	"testing"
)

func TestLatestHeight(t *testing.T) {
    height, err := GetLatestHeight()
    if height.Number < 1000 {
        t.Errorf(`GetLatestHeight = %v, %v`, height, err)
    }
}

func TestVaultAddress(t *testing.T) {
    bytes, _ := hex.DecodeString("02c72d6f1a74d169ddbdf5b7da258ece5fa09cc6b13385a8b0bcd7b1aef3bf4483")
    _, err := GetVaultAddress(bytes)
    if err != nil {
        t.Errorf(`TestVaultAddress = %v`, err)
    }
}

func TestValidateAddress(t *testing.T) {
    valid, err := ValidateAddress("t1ev8Fuh8t1bqheZZa7974j5jwKCjVcP7Pq")
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
    valid, err := MatchWithBlockchainReceiver("t1ev8Fuh8t1bqheZZa7974j5jwKCjVcP7Pq", "t1ev8Fuh8t1bqheZZa7974j5jwKCjVcP7Pq")
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
    balance, err := GetBalance("t1RyCw14wRXrh3mp21uxgr9ynjem7cNUkMH")
    if err != nil {
        t.Errorf(`TestBalance = %v`, err)
    }
    if balance == 0 {
        t.Errorf("This address should have some funds")
    }
}

