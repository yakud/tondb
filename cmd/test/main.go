package main

import (
	"fmt"

	"gitlab.flora.loc/mills/tondb/internal/utils"
)

func main() {
	//fmt.Println(utils.DecToHex(0))
	fmt.Println(utils.HexToDec("1000000000000000")) // 1152921504606846976
	//fmt.Println(0 << 3)                             // 2882303761517117440

	var shard_prefix = uint64(2305843009213693952)
	var shard_pfx_bits = uint64(4)
	var shard = shard_prefix | (1 << (63 - shard_pfx_bits))

	//
	//var shard = (shard_prefix & ((1 << (63 - shard_pfx_bits)) - 1)) == 0
	fmt.Println(shard)
	fmt.Println(utils.DecToHex(shard))
	//fmt.Println(utils.DecToHex(uint64(shard)))
}
