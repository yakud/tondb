package feed

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.flora.loc/mills/tondb/internal/ton/view"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	DefaultTransactionsLimit = 50
	MaxTransactionsLimit     = 500

	createTransactionsFeed = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_TransactionsFeed
	ENGINE = MergeTree() 
	PARTITION BY toYYYYMM(Time)
	ORDER BY (Time, Lt, MsgInCreatedLt, WorkchainId)
	SETTINGS index_granularity=128,index_granularity_bytes=0
	POPULATE 
	AS
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		Lt,
		Time,

	    arrayFilter((i, dir) -> dir = 'in',
	        arrayEnumerate(Messages.Direction),
	        Messages.Direction)[1] as MsgInIndex,
	       	
		Hash AS TrxHash,
	    Type,
	    Messages.CreatedLt[MsgInIndex]       AS MsgInCreatedLt,
	    Messages.Type[MsgInIndex]            AS MsgInType,
		Messages.SrcWorkchainId[MsgInIndex]  AS SrcWorkchainId, 
        Messages.SrcAddr[MsgInIndex]         AS Src, 
		Messages.DestWorkchainId[MsgInIndex] AS DestWorkchainId, 
        Messages.DestAddr[MsgInIndex]        AS Dest, 
		TotalFeesNanograms,
	    arraySum(Messages.ValueNanograms) as TotalNanograms,
	    arraySum(Messages.FwdFeeNanograms) as TotalFwdFeeNanograms,
	    arraySum(Messages.IhrFeeNanograms) as TotalIhrFeeNanograms,
	    arraySum(Messages.ImportFeeNanograms) as TotalImportFeeNanograms
	FROM transactions
`
	dropTransactionsFeed = `DROP TABLE _view_feed_TransactionsFeed`

	queryTransactionsFeedPart = `
	WITH (
		SELECT (min(Time), max(Time), max(Lt), max(MsgInCreatedLt))
		FROM (
			SELECT 
			   Time,
			   Lt,
			   MsgInCreatedLt
			FROM ".inner._view_feed_TransactionsFeed"
			PREWHERE
				 if(:time == 0, 1,
					(Time = :time AND Lt <= :lt AND MsgInCreatedLt < :message_lt) OR
					(Time < :time)
				 ) AND 
				 if(:workchain_id == bitShiftLeft(toInt32(-1), 31), 1, WorkchainId = :workchain_id)
			ORDER BY Time DESC, Lt DESC, MsgInCreatedLt DESC, WorkchainId DESC
			LIMIT :limit
		)
	) as TimeRange
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		toUInt64(Time) as TimeUnix,
		Lt,
		MsgInCreatedLt,
		TrxHash,
		Type,
		MsgInType,
		SrcWorkchainId,
		Src,
		DestWorkchainId,
		Dest,
		TotalNanograms,
		TotalFeesNanograms,
		TotalFwdFeeNanograms,
		TotalIhrFeeNanograms,
		TotalImportFeeNanograms
	FROM ".inner._view_feed_TransactionsFeed"
	PREWHERE 
		 (Time >= TimeRange.1 AND Time <= TimeRange.2) AND
		 (Lt <= TimeRange.3 AND MsgInCreatedLt <= TimeRange.4)
	ORDER BY Time DESC, Lt DESC, MsgInCreatedLt DESC, WorkchainId DESC
	LIMIT :limit
`
)

type TransactionInFeed struct {
	WorkchainId             int32  `db:"WorkchainId" json:"workchain_id"`
	Shard                   uint64 `db:"Shard" json:"shard"`
	SeqNo                   uint64 `db:"SeqNo" json:"seq_no"`
	TimeUnix                uint64 `db:"TimeUnix" json:"time"`
	Lt                      uint64 `db:"Lt" json:"lt"`
	MsgInCreatedLt          uint64 `db:"MsgInCreatedLt" json:"msg_in_created_lt"`
	TrxHash                 string `db:"TrxHash" json:"trx_hash"`
	Type                    string `db:"Type" json:"type"`
	MsgInType               string `db:"MsgInType" json:"msg_in_type"`
	SrcUf                   string `db:"-" json:"src_uf"`
	SrcWorkchainId          int32  `db:"SrcWorkchainId" json:"src_workchain_id"`
	Src                     string `db:"Src" json:"src"`
	DestUf                  string `db:"-" json:"dest_uf"`
	DestWorkchainId         int32  `db:"DestWorkchainId" json:"dest_workchain_id"`
	Dest                    string `db:"Dest" json:"dest"`
	TotalNanograms          uint64 `db:"TotalNanograms" json:"total_nanograms"`
	TotalFeesNanograms      uint64 `db:"TotalFeesNanograms" json:"total_fees_nanograms"`
	TotalFwdFeeNanograms    uint64 `db:"TotalFwdFeeNanograms" json:"total_fwd_fee_nanograms"`
	TotalIhrFeeNanograms    uint64 `db:"TotalIhrFeeNanograms" json:"total_ihr_fee_nanograms"`
	TotalImportFeeNanograms uint64 `db:"TotalImportFeeNanograms" json:"total_import_fee_nanograms"`
}

type TransactionsFeedScrollId struct {
	Time        uint64 `json:"t"`
	Lt          uint64 `json:"l"`
	MessageLt   uint64 `json:"m"`
	WorkchainId int32  `json:"w"`
}

type transactionsFeedDbFilter struct {
	Time        uint64 `db:"time"`
	Lt          uint64 `db:"lt"`
	MessageLt   uint64 `db:"message_lt"`
	Limit       uint16 `db:"limit"`
	WorkchainId int32  `db:"workchain_id"`
}

type TransactionsFeed struct {
	view.View
	conn *sqlx.DB
}

func (t *TransactionsFeed) CreateTable() error {
	_, err := t.conn.Exec(createTransactionsFeed)
	return err
}

func (t *TransactionsFeed) DropTable() error {
	_, err := t.conn.Exec(dropTransactionsFeed)
	return err
}

func (t *TransactionsFeed) SelectTransactions(scrollId *TransactionsFeedScrollId, limit uint16) ([]*TransactionInFeed, *TransactionsFeedScrollId, error) {
	if scrollId == nil {
		scrollId = &TransactionsFeedScrollId{
			WorkchainId: EmptyWorkchainId,
		}
	}
	if scrollId.WorkchainId == -2 {
		scrollId.WorkchainId = EmptyWorkchainId
	}
	if limit == 0 {
		limit = DefaultTransactionsLimit
	}
	if limit > MaxTransactionsLimit {
		limit = MaxTransactionsLimit
	}

	filter := transactionsFeedDbFilter{
		Time:        scrollId.Time,
		Lt:          scrollId.Lt,
		MessageLt:   scrollId.MessageLt,
		Limit:       limit,
		WorkchainId: scrollId.WorkchainId,
	}

	rows, err := t.conn.NamedQuery(queryTransactionsFeedPart, &filter)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var feed []*TransactionInFeed
	for rows.Next() {
		trx := &TransactionInFeed{}
		if err := rows.StructScan(trx); err != nil {
			return nil, nil, err
		}
		if trx.Src != "" {
			if trx.SrcUf, err = utils.ComposeRawAndConvertToUserFriendly(trx.SrcWorkchainId, trx.Src); err != nil {
				log.Println("src string:\"", trx.Src, "\"")
				log.Println("src bytes:\"", []byte(trx.Src), "\"")
				// Maybe we shouldn't fail here?
				return nil, nil, fmt.Errorf("error make uf address src: %w", err)
			}
		}
		if trx.Dest != "" {
			if trx.DestUf, err = utils.ComposeRawAndConvertToUserFriendly(trx.DestWorkchainId, trx.Dest); err != nil {
				// Maybe we shouldn't fail here?
				return nil, nil, fmt.Errorf("error make uf address dest: %w", err)
			}
		}

		feed = append(feed, trx)
	}

	if len(feed) == 0 {
		return feed, nil, nil
	}
	var lastTrx = feed[len(feed)-1]
	newScrollId := &TransactionsFeedScrollId{
		Time:        lastTrx.TimeUnix,
		Lt:          lastTrx.Lt,
		MessageLt:   lastTrx.MsgInCreatedLt,
		WorkchainId: scrollId.WorkchainId,
	}

	return feed, newScrollId, nil
}

func NewTransactionsFeed(conn *sqlx.DB) *TransactionsFeed {
	return &TransactionsFeed{
		conn: conn,
	}
}
