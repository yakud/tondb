package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	src := []byte("925D42FC88D29717a09BD5669313b684B45FDd9b")

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", dst[:n])
	fmt.Printf("%s\n", string(dst[:n]))
}
