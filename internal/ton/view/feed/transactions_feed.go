package feed

import (
	"fmt"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
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
	   	AccountAddr,
	   	IsTock,
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
		AccountAddr,
		IsTock,
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
		 (Lt <= TimeRange.3 AND MsgInCreatedLt <= TimeRange.4) AND 
		 if(:workchain_id == bitShiftLeft(toInt32(-1), 31), 1, WorkchainId = :workchain_id)
	ORDER BY Time DESC, Lt DESC, MsgInCreatedLt DESC, WorkchainId DESC
	LIMIT :limit
`
)

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

func (t *TransactionsFeed) SelectTransactions(scrollId *TransactionsFeedScrollId, limit uint16) ([]tonapi.TransactionsFeed, *TransactionsFeedScrollId, error) {
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
	var feed []tonapi.TransactionsFeed
	for rows.Next() {
		trx := tonapi.TransactionsFeed{}

		err := rows.Scan(&trx.WorkchainId, &trx.Shard, &trx.SeqNo, &trx.Time, &trx.Lt, &trx.MsgInCreatedLt, &trx.TrxHash,
			&trx.Type, &trx.AccountAddr, &trx.IsTock, &trx.MsgInType, &trx.SrcWorkchainId, &trx.Src, &trx.DestWorkchainId,
			&trx.Dest, &trx.TotalNanograms, &trx.TotalFeesNanograms, &trx.TotalFwdFeeNanograms, &trx.TotalIhrFeeNanograms,
			&trx.TotalImportFeeNanograms)

		if err != nil {
			return nil, nil, err
		}

		trx.Src = utils.NullAddrToString(trx.Src)
		trx.Dest = utils.NullAddrToString(trx.Dest)

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
		if trx.AccountAddr != "" {
			if trx.AccountAddrUf, err = utils.ComposeRawAndConvertToUserFriendly(trx.WorkchainId, trx.AccountAddr); err != nil {
				// Maybe we shouldn't fail here?
				return nil, nil, fmt.Errorf("error make uf address: %w", err)
			}
		}

		feed = append(feed, trx)
	}

	if len(feed) == 0 {
		return feed, nil, nil
	}
	var lastTrx = feed[len(feed)-1]
	newScrollId := &TransactionsFeedScrollId{
		Time:        uint64(lastTrx.Time),
		Lt:          uint64(lastTrx.Lt),
		MessageLt:   uint64(lastTrx.MsgInCreatedLt),
		WorkchainId: scrollId.WorkchainId,
	}

	return feed, newScrollId, nil
}

func NewTransactionsFeed(conn *sqlx.DB) *TransactionsFeed {
	return &TransactionsFeed{
		conn: conn,
	}
}
