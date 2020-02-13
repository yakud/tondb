package utils

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/npat-efault/crc16"
)

/*
Under the conditions stated above, the smart-contract address can be represented in the following forms:

A) "Raw": <decimal workchain_id>:<64 hexadecimal digits with address>
B) "User-friendly", which is obtained by first generating:
- one tag byte (0x11 for "bounceable" addresses, 0x51 for "non-bounceable"; add +0x80 if the address should not be accepted by software running in the production network)
- one byte containing a signed 8-bit integer with the workchain_id (0x00 for the basic workchain, 0xff for the masterchain)
- 32 bytes containing 256 bits of the smart-contract address inside the workchain (big-endian)
- 2 bytes containing CRC16-CCITT of the previous 34 bytes

In case B), the 36 bytes thus obtained are then encoded using base64 (i.e., with digits, upper- and lowercase Latin letters, '/' and '+') or base64url (with '_' and '-' instead of '/' and '+'), yielding 48 printable non-space characters.

Example:

The "test giver" (a special smart contract residing in the masterchain of the Test Network that gives up to 20 test Grams to anybody who asks) has the address

-1:fcb91a3a3816d0f7b8c2c76108b8a9bc5a6b7a55bd79f8ab101c52db29232260

in the "raw" form (notice that uppercase Latin letters 'A'..'F' may be used instead of 'a'..'f')

and

kf/8uRo6OBbQ97jCx2EIuKm8Wmt6Vb15-KsQHFLbKSMiYIny (base64)
kf_8uRo6OBbQ97jCx2EIuKm8Wmt6Vb15-KsQHFLbKSMiYIny (base64url)
*/

const (
	AddrTagBounceable          = 0x11
	AddrTagNonBounceable       = 0x51
	AddrTagDebugAddr           = 0x80
	UserFriendlyAddrDefaultTag = AddrTagBounceable

	Workchain0Byte  = 0x00
	MasterchainByte = 0xff

	addrRawBytesLength          = 32
	crcHashBytes                = 34
	addrUserFriendlyBytesLength = 36
)

var defaultBase64 = base64.RawURLEncoding

func ComposeRawAndConvertToUserFriendly(wcId int32, addr string) (string, error) {
	return ConvertRawToUserFriendly(strconv.Itoa(int(wcId))+":"+addr, UserFriendlyAddrDefaultTag)
}

func ConvertRawToUserFriendly(rawAddr string, tag byte) (string, error) {
	wid, addr, err := parseAccountAddressRaw(rawAddr)
	if err != nil {
		return "", err
	}

	addrBytes := make([]byte, hex.DecodedLen(len(addr)))
	addrBytesDecoded, err := hex.Decode(addrBytes, []byte(addr))
	if err != nil {
		return "", err
	}

	if addrBytesDecoded != addrRawBytesLength {
		return "", fmt.Errorf("addr should be exactly %d bytes actual: %d", addrRawBytesLength, addrBytesDecoded)
	}

	addrUfBytes := make([]byte, addrUserFriendlyBytesLength)

	// set the tag
	addrUfBytes[0] = tag

	// set the workchain_id
	switch wid {
	case -1:
		addrUfBytes[1] = MasterchainByte
	case 0:
		addrUfBytes[1] = Workchain0Byte
	}

	// set addr bytes
	copy(addrUfBytes[2:], addrBytes[:addrRawBytesLength])

	checksum := crc16.Checksum(crc16.XModem, addrUfBytes[:crcHashBytes])

	// crc16 put
	addrUfBytes[34] = byte(checksum >> 8)
	addrUfBytes[35] = byte(checksum & 0xff)

	return defaultBase64.EncodeToString(addrUfBytes), nil
}

func ConvertUserFriendlyToRaw(ufAddr string) (string, error) {
	if wc, addrHex, err := parseAccountAddressUserFriendly(ufAddr); err != nil {
		return "", err
	} else {
		return strconv.FormatInt(int64(wc), 10) + ":" + addrHex, nil
	}
}

var rawAddrMatcher = regexp.MustCompile(`[-]?\d:\S{64}`)

func ParseAccountAddress(addr string) (int32, string, error) {
	var err error
	addr, err = url.QueryUnescape(addr)
	if err != nil {
		return 0, "", err
	}

	match := rawAddrMatcher.MatchString(addr)

	if match {
		return parseAccountAddressRaw(addr)
	} else {
		return parseAccountAddressUserFriendly(addr)
	}
}

func parseAccountAddressRaw(addr string) (int32, string, error) {
	var err error

	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return 0, "", errors.New("wrong addr format. Should be workchainId:addrHash")
	}

	workchainId, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return 0, "", err
	}

	WorkchainId := int32(workchainId)
	Addr := strings.ToUpper(parts[1])

	return WorkchainId, Addr, nil
}

func parseAccountAddressUserFriendly(addr string) (int32, string, error) {
	addrUfBytes := make([]byte, addrUserFriendlyBytesLength)
	if _, err := defaultBase64.Decode(addrUfBytes[:addrUserFriendlyBytesLength], []byte(addr)); err != nil {
		return 0, "", err
	}

	checksum := crc16.Checksum(crc16.XModem, addrUfBytes[:crcHashBytes])

	if addrUfBytes[34] != uint8(checksum>>8) || addrUfBytes[35] != uint8(checksum&0xff) {
		return 0, "", fmt.Errorf("mismatch checksum")
	}

	if (addrUfBytes[0] & 0x3f) != 0x11 {
		return 0, "", fmt.Errorf("mismatch first byte")
	}

	wc := int32(int8(addrUfBytes[1]))
	addrHex := strings.ToUpper(hex.EncodeToString(addrUfBytes[2:34]))

	return wc, addrHex, nil
}
