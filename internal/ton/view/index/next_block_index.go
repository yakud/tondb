package index

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	createIndexNextBlock = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_index_NextBlock
	ENGINE = MergeTree()
	ORDER BY (WorkchainId, Shard, SeqNo)
	SETTINGS index_granularity = 64
	POPULATE 
	AS
	SELECT 
   		WorkchainId,
   		Shard,
   		SeqNo as NextSeqNo,
   		SeqNo-1 as SeqNo
	FROM blocks;
`
	dropIndexNextBlock = `DROP TABLE _view_index_NextBlock`
)

type IndexNextBlock struct {
	view.View
	conn *sql.DB
}

func (t *IndexNextBlock) CreateTable() error {
	_, err := t.conn.Exec(createIndexNextBlock)
	return err
}

func (t *IndexNextBlock) DropTable() error {
	_, err := t.conn.Exec(dropIndexNextBlock)
	return err
}
func NewIndexNextBlock(conn *sql.DB) *IndexNextBlock {
	return &IndexNextBlock{
		conn: conn,
	}
}
