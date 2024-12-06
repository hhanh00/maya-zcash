package main

import (
	"encoding/hex"
	"fmt"
	"mayazcash_go/maya_zcash"
)

func main() {
    // Call the Rust function
    height, err := maya_zcash.GetLatestHeight()
    if err != nil {
        fmt.Println(err);
    }

    fmt.Println(height.Number)
    fmt.Println(height.Hash)

    bytes, err := hex.DecodeString("02c72d6f1a74d169ddbdf5b7da258ece5fa09cc6b13385a8b0bcd7b1aef3bf4483")
    addr, err := maya_zcash.GetVaultAddress(bytes)
    fmt.Println(addr)
}
