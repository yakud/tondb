package storage

import (
	"database/sql"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

const (
	queryCreateTableAccountState string = `CREATE TABLE IF NOT EXISTS account_state (
		WorkchainId Int32,
		Shard       UInt64,
		SeqNo       UInt64,
		RootHash    FixedString(64),
		FileHash    FixedString(64),
		Time                   DateTime,
		Addr                   FixedString(64),
		Anycast                LowCardinality(String),
		Status                 String,
		BalanceNanogram        UInt64,
		Tick                   UInt64,
		Tock                   UInt64,
		StorageUsedBits        UInt64,
		StorageUsedCells       UInt64,
		StorageUsedPublicCells UInt64,
		LastTransHash          String,
		LastTransLt            UInt64,
		LastTransLtStorage     UInt64,
		LastPaid               UInt64
	) ENGINE MergeTree
	PARTITION BY (WorkchainId, Shard, round(SeqNo / 2000000))
	ORDER BY (WorkchainId, Addr, Shard, SeqNo);
`

	queryInsertAccountState = `INSERT INTO account_state (
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
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	queryDropAccountState = `DROP TABLE account_state;`
)

type AccountState struct {
	conn *sql.DB
}

func (s *AccountState) CreateTable() error {
	bdTx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	if _, err := bdTx.Exec(queryCreateTableAccountState); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *AccountState) DropTable() error {
	bdTx, err := s.conn.Begin()
	if err != nil {
		return err
	}

	if _, err := bdTx.Exec(queryDropAccountState); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *AccountState) InsertMany(states []*ton.AccountState) error {
	bdTx, err := s.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := s.InsertManyExec(states, bdTx)
	if err != nil {
		if stmt != nil {
			stmt.Close()
		}
		return err
	}

	if err := bdTx.Commit(); err != nil {
		if stmt != nil {
			stmt.Close()
		}
		return err
	}
	stmt.Close()

	return nil
}

func (s *AccountState) InsertManyExec(states []*ton.AccountState, bdTx *sql.Tx) (*sql.Stmt, error) {
	stmt, err := bdTx.Prepare(queryInsertAccountState)
	if err != nil {
		return stmt, err
	}

	for _, st := range states {
		// in order like BlocksFields
		if _, err := stmt.Exec(
			st.BlockId.WorkchainId,
			st.Shard,
			st.SeqNo,
			strings.TrimLeft(st.RootHash, "x"),
			strings.TrimLeft(st.FileHash, "x"),
			st.Time,
			strings.TrimLeft(st.Addr, "x"),
			st.Anycast,
			st.Status,
			st.BalanceNanogram,
			st.Tick,
			st.Tock,
			st.StorageUsedBits,
			st.StorageUsedCells,
			st.StorageUsedPublicCells,
			strings.TrimLeft(st.LastTransHash, "x"),
			st.LastTransLt,
			st.LastTransLtStorage,
			st.LastPaid,
		); err != nil {
			return stmt, err
		}
	}

	return stmt, nil
}

func NewAccountState(conn *sql.DB) *AccountState {
	s := &AccountState{
		conn: conn,
	}

	return s
}
