package utils

import (
	"strings"
)

// Truncate right zeros in decimal string
// from 0.010000 to 0.01
func TruncateRightZeros(value string) string {
	if !strings.Contains(value, ".") {
		return value
	}

	for i := len(value) - 1; i > 0; i-- {
		if value[i] == '0' {
			continue
		}
		if value[i] == '.' {
			return value[0:i]
		}

		return value[0 : i+1]
	}

	return value
}
