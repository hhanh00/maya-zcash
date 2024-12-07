package main

import (
	"encoding/hex"
	"fmt"
	"mayazcash_go/maya_zcash"
)

func main() {
    maya_zcash.InitLogger();

    // Call the Rust function
    height, err := maya_zcash.GetLatestHeight()
    if err != nil {
        fmt.Println(err);
    }

    fmt.Println(height.Number)
    fmt.Println(height.Hash)

    bytes, _ := hex.DecodeString("02c72d6f1a74d169ddbdf5b7da258ece5fa09cc6b13385a8b0bcd7b1aef3bf4483")
    addr, _ := maya_zcash.GetVaultAddress(bytes)
    fmt.Println(addr)

    valid, _ := maya_zcash.ValidateAddress("t1ev8Fuh8t1bqheZZa7974j5jwKCjVcP7Pq")
    fmt.Println(valid)

    valid, _ = maya_zcash.ValidateAddress("t1invalidaddress")
    fmt.Println(valid)

    // TODO Add more test cases
    valid, _ = maya_zcash.MatchWithBlockchainReceiver("t1ev8Fuh8t1bqheZZa7974j5jwKCjVcP7Pq", "t1ev8Fuh8t1bqheZZa7974j5jwKCjVcP7Pq")
    fmt.Println(valid)

    balance, _ := maya_zcash.GetBalance("t1RyCw14wRXrh3mp21uxgr9ynjem7cNUkMH")
    fmt.Println(balance)
}
