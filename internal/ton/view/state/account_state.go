package state

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton/view"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	queryGetAccount = `
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		RootHash,
		FileHash,
		Time,
		Addr,
		Anycast,
		Status,
		BalanceNanogram,
		Tick,
		Tock,
		StorageUsedBits,
		StorageUsedCells,
		StorageUsedPublicCells,
		LastTransHash,
		LastTransLt,
		LastTransLtStorage,
		LastPaid
	FROM ".inner._view_state_AccountState" FINAL
	WHERE %s
`
)

type AccountState struct {
	view.View
	conn *sql.DB
}

func (t *AccountState) GetAccount(f filter.Filter) (*ton.AccountState, error) {
	query, args, err := filter.RenderQuery(queryGetAccount, f)
	if err != nil {
		return nil, err
	}

	res := &ton.AccountState{}
	row := t.conn.QueryRow(query, args...)
	err = row.Scan(
		&res.BlockId.WorkchainId,
		&res.BlockId.Shard,
		&res.BlockId.SeqNo,
		&res.RootHash,
		&res.FileHash,
		&res.Time,
		&res.Addr,
		&res.Anycast,
		&res.Status,
		&res.BalanceNanogram,
		&res.Tick,
		&res.Tock,
		&res.StorageUsedBits,
		&res.StorageUsedCells,
		&res.StorageUsedPublicCells,
		&res.LastTransHash,
		&res.LastTransLt,
		&res.LastTransLtStorage,
		&res.LastPaid,
	)
	if err != nil {
		return nil, err
	}

	if res.AddrUf, err = utils.ComposeRawAndConvertToUserFriendly(res.WorkchainId, res.Addr); err != nil {
		// maybe we don't need to fail, just return account without user friendly address?
		return nil, err
	}

	return res, nil
}

func NewAccountState(conn *sql.DB) *AccountState {
	return &AccountState{
		conn: conn,
	}
}
