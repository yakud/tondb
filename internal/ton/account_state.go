package ton

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AccountState struct {
	BlockId
	BlockHeader

	Addr                   string    `json:"addr"`
	Time                   time.Time `json:"time"`
	Anycast                string    `json:"anycast"`
	Status                 string    `json:"status"`
	BalanceNanogram        uint64    `json:"balance_nanogram"`
	Tick                   uint64    `json:"tick"`
	Tock                   uint64    `json:"tock"`
	StorageUsedBits        uint64    `json:"storage_used_bits"`
	StorageUsedCells       uint64    `json:"storage_used_cells"`
	StorageUsedPublicCells uint64    `json:"storage_used_public_cells"`
	LastTransHash          string    `json:"last_trans_hash"`
	LastTransLt            uint64    `json:"last_trans_lt"`
	LastTransLtStorage     uint64    `json:"last_trans_lt_storage"`
	LastPaid               uint64    `json:"last_paid"`
}

func ParseAccountAddress(addr string) (AddrStd, error) {
	var err error
	var addrStd AddrStd

	addr, err = url.QueryUnescape(addr)
	if err != nil {
		return addrStd, err
	}

	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return addrStd, errors.New("wrong addr format. Should be workchainId:addrHash")
	}

	workchainId, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return addrStd, err
	}

	addrStd.WorkchainId = int32(workchainId)
	addrStd.Addr = strings.ToUpper(parts[1])

	return addrStd, nil
}
