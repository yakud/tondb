package utils

import (
	"strings"
	"testing"
)

/*

new wallet address = -1:acde1b6b54cfc041e9d473ef2b7d9f175b86904db3c58112505a2b5a1921ce34
(Saving address to file neonbones_master.addr)
Non-bounceable address (for init): 0f-s3htrVM_AQenUc-8rfZ8XW4aQTbPFgRJQWitaGSHONJgG
Bounceable address (for later access): kf-s3htrVM_AQenUc-8rfZ8XW4aQTbPFgRJQWitaGSHONMXD
signing message: x{00000000FFFFFFFF}

account hex	ACDE1B6B54CFC041E9D473EF2B7D9F175B86904DB3C58112505A2B5A1921CE34
account	Ef-s3htrVM_AQenUc-8rfZ8XW4aQTbPFgRJQWitaGSHONH5J

*/
func TestConvertRawToUserFriendly(t *testing.T) {
	addrs := map[string]string{
		"-1:acde1b6b54cfc041e9d473ef2b7d9f175b86904db3c58112505a2b5a1921ce34": "Ef-s3htrVM_AQenUc-8rfZ8XW4aQTbPFgRJQWitaGSHONH5J",
		"-1:34517C7BDF5187C55AF4F8B61FDC321588C7AB768DEE24B006DF29106458D7CF": "Ef80UXx731GHxVr0-LYf3DIViMerdo3uJLAG3ykQZFjXz2kW",
		"0:A0201CEA76F1185807624ED15BEE6E0290066CFEAAFAF57D2BD270817BEEFACB":  "EQCgIBzqdvEYWAdiTtFb7m4CkAZs_qr69X0r0nCBe-76y6Va",
	}
	for raw, uf := range addrs {
		ufAddr, err := ConvertRawToUserFriendly(raw, AddrTagBounceable)
		if err != nil {
			t.Fatal(err)
		}

		if ufAddr != uf {
			t.Fatalf("error expected: %s actual: %s", uf, ufAddr)
		}

		raw2, err := ConvertUserFriendlyToRaw(ufAddr)
		if err != nil {
			t.Fatal(err)
		}

		if strings.ToUpper(raw) != raw2 {
			t.Fatalf("error raw addr expected: %s actual: %s", raw, raw2)
		}

		raw2, err = ConvertUserFriendlyToRaw(uf)
		if err != nil {
			t.Fatal(err)
		}

		if strings.ToUpper(raw) != raw2 {
			t.Fatalf("error raw addr expected: %s actual: %s", raw, raw2)
		}
	}
}
