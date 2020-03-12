package feed

import (
	"log"

	"github.com/jmoiron/sqlx"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	DefaultBlocksLimit = 30
	MaxBlocksLimit     = 500

	// todo: recreate table
	createBlocksFeed = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_BlocksFeed
	ENGINE = MergeTree()
	PARTITION BY toStartOfYear(Time)
	ORDER BY (Time, StartLt, Shard, WorkchainId)
	SETTINGS index_granularity=64,index_granularity_bytes=0
	POPULATE
	AS
	SELECT
		WorkchainId,
		Shard,
		SeqNo,
		Time,
		StartLt,
		RootHash,
	    FileHash,
		BlockStatsTrxCount,
		BlockStatsSentNanograms,
		ValueFlowFeesCollected
	FROM blocks
`

	// todo: recreate table
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
		count() AS TrxCount,
		sumArray(Messages.ValueNanograms) AS ValueNanograms,
		sumArray(Messages.IhrFeeNanograms) AS IhrFeeNanograms,
		sumArray(Messages.ImportFeeNanograms) AS ImportFeeNanograms,
		sumArray(Messages.FwdFeeNanograms) AS FwdFeeNanograms
	FROM transactions
	GROUP BY Time, TotalFeesNanograms, WorkchainId, Shard, SeqNo
`

	dropBlocksFeed = `DROP TABLE _view_feed_BlocksFeed`

	dropTransactionFeesFeed = `DROP TABLE _view_feed_TransactionFeesFeed`

	queryBlocksFeedScoll = `
	WITH (
		SELECT (min(Time), max(Time), max(StartLt), max(Shard))
		FROM (
			SELECT 
			   Time,
			   StartLt,
			   Shard
			FROM ".inner._view_feed_BlocksFeed"
			PREWHERE
				 if(:time == 0, 1,
					(Time = :time AND StartLt <= :lt AND Shard < :shard) OR
					(Time < :time)
				 ) AND 
				 if(:workchain_id == bitShiftLeft(toInt32(-1), 31), 1, WorkchainId = :workchain_id)
			ORDER BY Time DESC, StartLt DESC, Shard DESC, WorkchainId DESC
			LIMIT :limit
		)
	) as TimeRange
	SELECT
		 WorkchainId,
		 Shard,
		 SeqNo,
		 toUInt64(Time) as TimeUnix,
	     StartLt,
	     RootHash,
	     FileHash,
	     BlockStatsTrxCount as TrxCount,
		 BlockStatsSentNanograms as ValueNanograms,
		 ValueFlowFeesCollected as TotalFeesNanograms
	 FROM ".inner._view_feed_BlocksFeed"
	 PREWHERE
		 (Time >= TimeRange.1 AND Time <= TimeRange.2)  AND
		 (StartLt <= TimeRange.3 AND Shard <= TimeRange.4) AND
		 if(:workchain_id != bitShiftLeft(toInt32(-1), 31), WorkchainId = :workchain_id, 1)
	 ORDER BY Time DESC, StartLt DESC, Shard DESC, WorkchainId DESC
`
)

type BlockInFeed struct {
	WorkchainId        int32  `db:"WorkchainId" json:"workchain_id"`
	Shard              uint64 `db:"Shard" json:"shard"`
	SeqNo              uint64 `db:"SeqNo" json:"seq_no"`
	Time               uint64 `db:"TimeUnix" json:"time"`
	StartLt            uint64 `db:"StartLt" json:"start_lt"`
	RootHash           string `db:"RootHash" json:"root_hash"`
	FileHash           string `db:"FileHash" json:"file_hash"`
	TotalFeesNanograms uint64 `db:"TotalFeesNanograms" json:"total_fees_nanograms"`
	TrxCount           uint64 `db:"TrxCount" json:"trx_count"`
	ValueNanograms     uint64 `db:"ValueNanograms" json:"value_nanograms"`
}

type BlocksFeedScrollId struct {
	Time        uint64 `json:"t"`
	Lt          uint64 `json:"l"`
	Shard       uint64 `json:"m"`
	WorkchainId int32  `json:"w"`
}

type blocksFeedDbFilter struct {
	Time        uint64 `db:"time"`
	Lt          uint64 `db:"lt"`
	Shard       uint64 `db:"shard"`
	Limit       uint16 `db:"limit"`
	WorkchainId int32  `db:"workchain_id"`
}

type BlocksFeed struct {
	view.View
	conn *sqlx.DB
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

func (t *BlocksFeed) SelectBlocks(scrollId *BlocksFeedScrollId, limit uint16) ([]*BlockInFeed, *BlocksFeedScrollId, error) {
	if scrollId == nil {
		scrollId = &BlocksFeedScrollId{
			WorkchainId: EmptyWorkchainId,
		}
	}
	if scrollId.WorkchainId == -2 {
		scrollId.WorkchainId = EmptyWorkchainId
	}
	if limit == 0 {
		limit = DefaultBlocksLimit
	}
	if limit > MaxBlocksLimit {
		limit = MaxBlocksLimit
	}

	filter := blocksFeedDbFilter{
		Time:        scrollId.Time,
		Lt:          scrollId.Lt,
		Shard:       scrollId.Shard,
		Limit:       limit,
		WorkchainId: scrollId.WorkchainId,
	}

	rows, err := t.conn.NamedQuery(queryBlocksFeedScoll, &filter)
	if err != nil {
		log.Println("fetch blocks err:", err)
		return nil, nil, err
	}
	defer rows.Close()

	var feed []*BlockInFeed
	for rows.Next() {
		msg := &BlockInFeed{}
		if err := rows.StructScan(msg); err != nil {
			return nil, nil, err
		}

		feed = append(feed, msg)
	}

	if len(feed) == 0 {
		return feed, nil, nil
	}

	var lastMsg = feed[len(feed)-1]
	newScrollId := &BlocksFeedScrollId{
		Time:        lastMsg.Time,
		Lt:          lastMsg.StartLt,
		Shard:       lastMsg.Shard,
		WorkchainId: scrollId.WorkchainId,
	}

	return feed, newScrollId, nil
}

func NewBlocksFeed(conn *sqlx.DB) *BlocksFeed {
	return &BlocksFeed{
		conn: conn,
	}
}
