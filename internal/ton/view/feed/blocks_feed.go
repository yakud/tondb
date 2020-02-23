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
		Time
	FROM blocks
`

	createTransactionFeesFeed = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_TransactionFeesFeed
	ENGINE = SummingMergeTree()  
	PARTITION BY toStartOfYear(Time)
	ORDER BY (Time, WorkchainId, Shard, SeqNo)
	SETTINGS index_granularity=128,index_granularity_bytes=0
	POPULATE 
	AS
	SELECT 
	    Time,   
		TotalFeesNanograms,
		WorkchainId,
		Shard,
		SeqNo,
		count() AS Count,
		sumArray(Messages.ValueNanograms) AS ValueNanograms,
		sumArray(Messages.IhrFeeNanograms) AS IhrFeeNanograms,
		sumArray(Messages.ImportFeeNanograms) AS ImportFeeNanograms,
		sumArray(Messages.FwdFeeNanograms) AS FwdFeeNanograms
	FROM transactions
	GROUP BY Time, TotalFeesNanograms, WorkchainId, Shard, SeqNo
`

	dropBlocksFeed = `DROP TABLE _view_feed_BlocksFeed`

	dropTransactionFeesFeed = `DROP TABLE _view_feed_TransactionFeesFeed`

	// TODO: optimize query make it beauty
	queryBlocksFeed = `
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		toUInt64(Time),
	    TotalFeesNanograms,
	    Count,
	    ValueNanograms,
	    IhrFeeNanograms,
	    ImportFeeNanograms,
	    FwdFeeNanograms
	FROM (
	    WITH (
			SELECT (min(Time), max(Time))
			FROM (
				SELECT 
				   Time
				FROM ".inner._view_feed_BlocksFeed"
				PREWHERE
					 if(? != 0, Time < toDateTime(?), 1) AND
					 if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId = ?, 1)
				ORDER BY Time DESC, WorkchainId DESC, Shard DESC, SeqNo DESC
				LIMIT ?
			)
		) as TimeRange
	     SELECT 
			 WorkchainId,
			 Shard,
			 SeqNo,
			 Time
	     FROM ".inner._view_feed_BlocksFeed"
	     PREWHERE 
			 (Time >= TimeRange.1 AND Time <= TimeRange.2) AND
			 if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId = ?, 1)
		 ORDER BY Time DESC, WorkchainId DESC, Shard DESC, SeqNo DESC
	) ANY LEFT JOIN ( 
		WITH (
			SELECT (min(Time), max(Time))
			FROM (
				SELECT 
				   Time
				FROM ".inner._view_feed_BlocksFeed"
				PREWHERE
					 if(? != 0, Time < toDateTime(?), 1) AND
					 if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId = ?, 1)
				ORDER BY Time DESC, WorkchainId DESC, Shard DESC, SeqNo DESC
				LIMIT ?
			)
		) as TimeRange
		SELECT
		    WorkchainId,
			Shard,
			SeqNo,
			TotalFeesNanograms,
			Count,
			ValueNanograms,
			IhrFeeNanograms,
			ImportFeeNanograms,
			FwdFeeNanograms
		FROM ".inner._view_feed_TransactionFeesFeed"
		PREWHERE 
			 (Time >= TimeRange.1 AND Time <= TimeRange.2) AND
			 if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId = ?, 1)
	) USING (WorkchainId, Shard, SeqNo);
`
)

type BlockInFeed struct {
	WorkchainId        int32  `json:"workchain_id"`
	Shard              uint64 `json:"shard"`
	SeqNo              uint64 `json:"seq_no"`
	Time               uint64 `json:"time"`
	TotalFeesNanograms uint64 `json:"total_fees_nanograms"`
	Count              uint64 `json:"count"`
	ValueNanograms     uint64 `json:"value_nanograms"`
	IhrFeeNanograms    uint64 `json:"ihr_fee_nanograms"`
	ImportFeeNanograms uint64 `json:"import_fee_nanograms"`
	FwdFeeNanograms    uint64 `json:"fwd_fee_nanograms"`
}

type BlocksFeed struct {
	view.View
	conn *sql.DB
}

func (t *BlocksFeed) CreateTable() error {
	_, err := t.conn.Exec(createBlocksFeed)
	_, err = t.conn.Exec(createTransactionFeesFeed)
	return err
}

func (t *BlocksFeed) DropTable() error {
	_, err := t.conn.Exec(dropBlocksFeed)
	_, err = t.conn.Exec(dropTransactionFeesFeed)
	return err
}

func (t *BlocksFeed) SelectBlocks(wcId int32, limit int16, beforeTime time.Time) ([]*BlockInFeed, error) {
	beforeTimeInt := beforeTime.Unix()
	rows, err := t.conn.Query(
		queryBlocksFeed,
		beforeTimeInt, beforeTimeInt, wcId, wcId, limit, wcId, wcId,
		beforeTimeInt, beforeTimeInt, wcId, wcId, limit, wcId, wcId,
	)
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
			&row.TotalFeesNanograms,
			&row.Count,
			&row.ValueNanograms,
			&row.IhrFeeNanograms,
			&row.ImportFeeNanograms,
			&row.FwdFeeNanograms,
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
