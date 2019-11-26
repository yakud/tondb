package utils

import (
	"testing"
)

func TestTruncateRightZeros(t *testing.T) {
	if a := TruncateRightZeros("0.0001000"); a != "0.0001" {
		t.Fatal("0.0001000 should be 0.0001 actual:", a)
	}
	if a := TruncateRightZeros("0.0001010"); a != "0.000101" {
		t.Fatal("0.0001010 should be 0.000101 actual:", a)
	}
	if a := TruncateRightZeros("10.000000"); a != "10" {
		t.Fatal("10.000000 should be 10 actual:", a)
	}
	if a := TruncateRightZeros("0.000000"); a != "0" {
		t.Fatal("0.000000 should be 0 actual:", a)
	}
	if a := TruncateRightZeros("0."); a != "0" {
		t.Fatal("0. should be 0 actual:", a)
	}
	if a := TruncateRightZeros("100."); a != "100" {
		t.Fatal("100. should be 100 actual:", a)
	}
	if a := TruncateRightZeros("0"); a != "0" {
		t.Fatal("0. should be 0 actual:", a)
	}
	if a := TruncateRightZeros("0.1"); a != "0.1" {
		t.Fatal("0.1 should be 0.1 actual:", a)
	}
	if a := TruncateRightZeros("50000"); a != "50000" {
		t.Fatal("50000 should be 50000 actual:", a)
	}
}
