package timeseries

import "database/sql"

const (
	createTsBlocksByWorkchain = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _ts_BlocksByWorkchain
	ENGINE = SummingMergeTree() 
	PARTITION BY tuple()
	ORDER BY (Time, WorkchainId)
	POPULATE 
	AS
	SELECT
		toStartOfInterval(Time, INTERVAL 5 SECOND) as Time,
		WorkchainId,
		count() as Blocks
	FROM 
		blocks
	GROUP BY Time, WorkchainId
`

	dropTsBlocksByWorkchain = `DROP TABLE _ts_BlocksByWorkchain`
)

type BlocksByWorkchain struct {
	conn *sql.DB
}

func (t *BlocksByWorkchain) CreateTable() error {
	_, err := t.conn.Exec(createTsBlocksByWorkchain)
	return err
}

func (t *BlocksByWorkchain) DropTable() error {
	_, err := t.conn.Exec(dropTsBlocksByWorkchain)
	return err
}

func NewBlocksByWorkchain(conn *sql.DB) *BlocksByWorkchain {
	return &BlocksByWorkchain{
		conn: conn,
	}
}
