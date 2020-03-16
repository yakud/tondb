package ton

import (
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

type AccountState struct {
	BlockId
	BlockHeader

	Addr                   string    `json:"addr"`
	AddrUf                 string    `json:"addr_uf"`
	Time                   uint64    `json:"time"`
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
	if wc, addrHex, err := utils.ParseAccountAddress(addr); err != nil {
		return AddrStd{}, err
	} else {
		return AddrStd{IsEmpty: false, WorkchainId: wc, Addr: addrHex}, nil
	}
}
