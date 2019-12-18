package feed

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
	"gitlab.flora.loc/mills/tondb/internal/utils"
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
			PREWHERE if(? != 0, Time < toDateTime(?), 1 == 1)
			ORDER BY Time DESC, WorkchainId DESC, Shard DESC, SeqNo DESC
			LIMIT ?
		)
	) as TimeRange
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		Time,
	    StartLt,
	    EndLt
	FROM ".inner._view_feed_BlocksFeed"
	PREWHERE 
		(Time >= TimeRange.1 AND Time <= TimeRange.2)
	ORDER BY Time DESC, WorkchainId DESC, Shard DESC, SeqNo DESC
`
)

type BlockStat struct {
	WorkchainId   int32  `json:"workchain_id"`
	Shard         string `json:"shard"`
	SeqNo         uint64 `json:"seq_no"`
	Lt            uint64 `json:"lt"`
	Time          uint64 `json:"time"`
	Direction     string `json:"direction"`
	Src           string `json:"src"`
	Dest          string `json:"dest"`
	ValueGrams    string `json:"value_grams"`
	TotalFeeGrams string `json:"total_fee_grams"`
	Bounce        bool   `json:"bounce"`
}

type BlocksStats struct {
	view.View
	conn *sql.DB
}

func (t *BlocksStats) CreateTable() error {
	_, err := t.conn.Exec(createFeedAccountTransactions)
	return err
}

func (t *BlocksStats) DropTable() error {
	_, err := t.conn.Exec(dropFeedAccountTransactions)
	return err
}

func (t *BlocksStats) SelectLatestMessages(count int) ([]*MessageFeedGlobal, error) {
	rows, err := t.conn.Query(querySelectLastNMessages, count)
	if err != nil {
		if rows != nil {
			rows.Close()
		}

		return nil, err
	}

	res := make([]*MessageFeedGlobal, 0, count)
	for rows.Next() {
		row := &MessageFeedGlobal{}
		err := rows.Scan(
			&row.WorkchainId,
			&row.Shard,
			&row.SeqNo,
			&row.Lt,
			&row.Time,
			&row.Direction,
			&row.Src,
			&row.Dest,
			&row.ValueGrams,
			&row.TotalFeeGrams,
			&row.Bounce,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}
		row.ValueGrams = utils.TruncateRightZeros(row.ValueGrams)
		row.TotalFeeGrams = utils.TruncateRightZeros(row.TotalFeeGrams)

		res = append(res, row)
	}

	if rows != nil {
		rows.Close()
	}

	return res, err
}

func NewBlocksStats(conn *sql.DB) *BlocksStats {
	return &BlocksStats{
		conn: conn,
	}
}
