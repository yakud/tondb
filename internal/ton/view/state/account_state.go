package state

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	createStateAccountState = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_state_AccountState
	ENGINE = ReplacingMergeTree(SeqNo)
	ORDER BY (WorkchainId, Addr)
	SETTINGS index_granularity = 64
	POPULATE 
	AS
	SELECT 
		*    
	FROM account_state;
`
	dropStateAccountState = `DROP TABLE _view_state_AccountState`

	queryGetAccount = `
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		RootHash,
		FileHash,
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
	FROM ".inner._view_state_AccountState"
	WHERE %s
`
)

type AccountState struct {
	view.View
	conn *sql.DB
}

func (t *AccountState) CreateTable() error {
	_, err := t.conn.Exec(createStateAccountState)
	return err
}

func (t *AccountState) DropTable() error {
	_, err := t.conn.Exec(dropStateAccountState)
	return err
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

	return res, nil
}

func NewAccountState(conn *sql.DB) *AccountState {
	return &AccountState{
		conn: conn,
	}
}
