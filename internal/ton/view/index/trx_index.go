package index

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	createIndexTransactionBlock = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_index_TransactionBlock
	ENGINE = ReplacingMergeTree()
	ORDER BY Hash
	SETTINGS index_granularity = 64
	POPULATE 
	AS
	SELECT 
		cityHash64(Hash) as Hash,
		WorkchainId,
		Shard,      
		SeqNo      
	FROM transactions;
`
	dropIndexTransactionBlock = `DROP TABLE _view_index_TransactionBlock`
)

type IndexTransactionBlock struct {
	view.View
	conn *sql.DB
}

func (t *IndexTransactionBlock) CreateTable() error {
	_, err := t.conn.Exec(createIndexTransactionBlock)
	return err
}

func (t *IndexTransactionBlock) DropTable() error {
	_, err := t.conn.Exec(dropIndexTransactionBlock)
	return err
}
func NewIndexTransactionBlock(conn *sql.DB) *IndexTransactionBlock {
	return &IndexTransactionBlock{
		conn: conn,
	}
}
