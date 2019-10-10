package utils

import (
	"strconv"
)

func BoolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func DecToHex(dec uint64) string {
	return strconv.FormatUint(dec, 16)
}

func HexToDec(hex string) (uint64, error) {
	return strconv.ParseUint(hex, 16, 64)
}
