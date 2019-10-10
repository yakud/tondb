package ch

import (
	"testing"
)

func TestConnect(t *testing.T) {
	//schema://user:password@host[:port]/database?param1=value1&...&paramN=valueN
	_, err := Connect(nil)
	if err != nil {
		t.Fatal(err)
	}
}
