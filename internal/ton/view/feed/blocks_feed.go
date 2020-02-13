package feed

import (
	"database/sql"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	createBlocksFeed = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_BlocksFeed
	ENGINE = MergeTree() 
	PARTITION BY toStartOfYear(Time)
	ORDER BY (Time, WorkchainId, Shard, SeqNo)
	SETTINGS index_granularity=128,index_granularity_bytes=0
	POPULATE 
	AS
	SELECT
		WorkchainId,
		Shard,
		SeqNo,
	   	Time,
	    StartLt,
	    EndLt
	FROM blocks
`

	dropBlocksFeed = `DROP TABLE _view_feed_BlocksFeed`

	queryBlocksFeed = `
	WITH (
		SELECT (min(Time), max(Time))
		FROM (
			SELECT 
			   Time
			FROM ".inner._view_feed_BlocksFeed"
			PREWHERE
			     if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId == ?, 1 == 1)
			     AND if(? != 0, Time < toDateTime(?), 1 == 1)
			ORDER BY Time DESC, WorkchainId DESC, Shard DESC, SeqNo DESC
			LIMIT ?
		)
	) as TimeRange
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		toUInt64(Time),
	    StartLt,
	    EndLt
	FROM ".inner._view_feed_BlocksFeed"
	PREWHERE 
		if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId == ?, 1 == 1)
		AND (Time >= TimeRange.1 AND Time <= TimeRange.2)
	ORDER BY Time DESC, WorkchainId DESC, Shard DESC, SeqNo DESC
`
)

type BlockInFeed struct {
	WorkchainId int32  `json:"workchain_id"`
	Shard       uint64 `json:"shard"`
	SeqNo       uint64 `json:"seq_no"`
	Time        uint64 `json:"time"`
	StartLt     uint64 `json:"start_lt"`
	EndLt       uint64 `json:"end_lt"`
}

type BlocksFeed struct {
	view.View
	conn *sql.DB
}

func (t *BlocksFeed) CreateTable() error {
	_, err := t.conn.Exec(createBlocksFeed)
	return err
}

func (t *BlocksFeed) DropTable() error {
	_, err := t.conn.Exec(dropBlocksFeed)
	return err
}

func (t *BlocksFeed) SelectBlocks(wcId int32, limit int16, beforeTime time.Time) ([]*BlockInFeed, error) {
	beforeTimeInt := beforeTime.Unix()
	rows, err := t.conn.Query(queryBlocksFeed, wcId, wcId, beforeTimeInt, beforeTimeInt, limit, wcId, wcId)
	if err != nil {
		if rows != nil {
			rows.Close()
		}

		return nil, err
	}

	res := make([]*BlockInFeed, 0, limit)
	for rows.Next() {
		row := &BlockInFeed{}
		err := rows.Scan(
			&row.WorkchainId,
			&row.Shard,
			&row.SeqNo,
			&row.Time,
			&row.StartLt,
			&row.EndLt,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}

		res = append(res, row)
	}

	if rows != nil {
		rows.Close()
	}

	return res, err
}

func NewBlocksFeed(conn *sql.DB) *BlocksFeed {
	return &BlocksFeed{
		conn: conn,
	}
}
