package main

import (
	"fmt"
	"mayazcash_go/fb"
	"mayazcash_go/maya_zcash"
)

func main() {
    // Call the Rust function
    heightBuffer, _ := maya_zcash.GetLatestHeight()
    height := fb.GetRootAsHeight(heightBuffer, 0);
    fmt.Println(height.Number())
    fmt.Println(height.HashBytes())
}
