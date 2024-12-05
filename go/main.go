package main

import (
	"fmt"
	"mayazcash_go/maya_zcash"
)

func main() {
    // Call the Rust function
    height, _ := maya_zcash.GetLatestHeight()
    fmt.Println(height.Number)
    fmt.Println(height.Hash)
}
