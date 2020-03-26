package state

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/view"
	"gitlab.flora.loc/mills/tondb/internal/utils"
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
		toUInt64(Time),
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
	PREWHERE WorkchainId = ? AND Addr = ?
`
	queryGetAccountWithStats = `
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		RootHash,
		FileHash,
		toUInt64(Time),
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
		LastPaid,
		(
			SELECT
				sum(MessagesCount)
			FROM ".inner._view_AccountMessagesCount" FINAL
			PREWHERE WorkchainId = ? AND AccountAddr = ?
		) as MessagesCount
	FROM ".inner._view_state_AccountState" FINAL
	PREWHERE WorkchainId = ? AND Addr = ?
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

func (t *AccountState) GetAccount(addr ton.AddrStd) (*tonapi.Account, error) {
	res := &tonapi.Account{}
	row := t.conn.QueryRow(queryGetAccount, addr.WorkchainId, addr.Addr)
	err := row.Scan(
		&res.WorkchainId,
		&res.Shard,
		&res.SeqNo,
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

	res.Addr = utils.NullAddrToString(res.Addr)

	if res.AddrUf, err = utils.ComposeRawAndConvertToUserFriendly(*res.WorkchainId, res.Addr); err != nil {
		// maybe we don't need to fail, just return account without user friendly address?
		return nil, err
	}

	return res, nil
}

func (t *AccountState) GetAccountWithStats(addr ton.AddrStd) (*tonapi.Account, error) {
	res := &tonapi.Account{}
	row := t.conn.QueryRow(queryGetAccountWithStats, addr.WorkchainId, addr.Addr, addr.WorkchainId, addr.Addr)
	err := row.Scan(
		&res.WorkchainId,
		&res.Shard,
		&res.SeqNo,
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
		&res.MessagesCount,
	)
	if err != nil {
		return nil, err
	}

	res.Addr = utils.NullAddrToString(res.Addr)

	if res.AddrUf, err = utils.ComposeRawAndConvertToUserFriendly(*res.WorkchainId, res.Addr); err != nil {
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
