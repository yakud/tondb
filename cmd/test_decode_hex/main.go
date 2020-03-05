package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	src := []byte("B76E8FFF77F662CB16CDEB19E21BC533F4B41A703C530DAD513A25700BCE33E676103E0D1BDBDE8BD71996CE4474B3522FD8E8401770308DC6C7C53D3678050700001FB2FFFFFFFF03")

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", dst[:n])
	fmt.Printf("%s\n", string(dst[:n]))
}
