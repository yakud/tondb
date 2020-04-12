package stats

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	createAccountMessagesCount = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_AccountMessagesCount
	ENGINE = SummingMergeTree() 
	PARTITION BY tuple()
	ORDER BY (WorkchainId, AccountAddr)
	SETTINGS index_granularity=8,index_granularity_bytes=0
	POPULATE 
	AS
	SELECT
		WorkchainId,
		AccountAddr,
		count() as MessagesCount
	FROM transactions
	ARRAY JOIN Messages
	GROUP BY WorkchainId, AccountAddr
`
	dropAccountMessagesCount = `DROP TABLE _view_AccountMessagesCount`

	querySelectAccountMessagesCount = `
	SELECT 
		sum(MessagesCount) 
	FROM ".inner._view_AccountMessagesCount" FINAL
	PREWHERE WorkchainId = ? AND AccountAddr = ?
`
)

type AccountMessagesCount struct {
	view.View
	conn *sql.DB
}

func (t *AccountMessagesCount) CreateTable() error {
	_, err := t.conn.Exec(createAccountMessagesCount)
	return err
}

func (t *AccountMessagesCount) DropTable() error {
	_, err := t.conn.Exec(dropAccountMessagesCount)
	return err
}

func NewAccountMessagesCount(conn *sql.DB) *AccountMessagesCount {
	return &AccountMessagesCount{
		conn: conn,
	}
}
