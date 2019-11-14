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
	bdTx, err := t.conn.Begin()
	if err != nil {
		return err
	}

	if _, err := bdTx.Exec(createTsBlocksByWorkchain); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (t *BlocksByWorkchain) DropTable() error {
	bdTx, err := t.conn.Begin()
	if err != nil {
		return err
	}

	if _, err := bdTx.Exec(dropTsBlocksByWorkchain); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func NewBlocksByWorkchain(conn *sql.DB) *BlocksByWorkchain {
	return &BlocksByWorkchain{
		conn: conn,
	}
}
