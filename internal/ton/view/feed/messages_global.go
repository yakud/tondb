package feed

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	createMessagesFeedGlobal = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_MessagesFeedGlobal
	ENGINE = MergeTree() 
	PARTITION BY tuple()
	ORDER BY (Time, Lt)
	SETTINGS index_granularity=128,index_granularity_bytes=0
	POPULATE 
	AS
	SELECT
		WorkchainId,
		Shard,
		SeqNo,
		Lt,
		Time,
	    Messages.Direction as Direction, 
		Messages.SrcAddr AS Src, 
		Messages.DestAddr AS Dest, 
		Messages.ValueNanograms as ValueNanograms,
	    Messages.FwdFeeNanograms + Messages.IhrFeeNanograms + Messages.ImportFeeNanograms as TotalFeeNanograms, 
		Messages.Bounce as Bounce
	FROM transactions
	ARRAY JOIN Messages
	WHERE Type = 'trans_ord' AND Messages.Type = 'int_msg_info'
`
	dropMessagesFeedGlobal = `DROP TABLE _view_feed_MessagesFeedGlobal`

	querySelectLastNMessages = `SELECT 
	WorkchainId,
	hex(Shard),
	SeqNo,
	Lt,
	toUInt64(Time),
	Direction,
	Src,
	Dest,
	toDecimal128(ValueNanograms, 10) * toDecimal128(0.000000001, 10) as ValueGrams,
	toDecimal128(TotalFeeNanograms, 10) * toDecimal128(0.000000001, 10) as TotalFeeGrams,
	Bounce
FROM ton2._view_feed_MessagesFeedGlobal
WHERE Dest != '3333333333333333333333333333333333333333333333333333333333333333'
ORDER BY Time DESC, Lt DESC
LIMIT ?
`
)

type MessageFeedGlobal struct {
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

type MessagesFeedGlobal struct {
	view.View
	conn *sql.DB
}

func (t *MessagesFeedGlobal) CreateTable() error {
	_, err := t.conn.Exec(createMessagesFeedGlobal)
	return err
}

func (t *MessagesFeedGlobal) DropTable() error {
	_, err := t.conn.Exec(dropMessagesFeedGlobal)
	return err
}

func (t *MessagesFeedGlobal) SelectLatestMessages(count int) ([]*MessageFeedGlobal, error) {
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

func NewMessagesFeedGlobal(conn *sql.DB) *MessagesFeedGlobal {
	return &MessagesFeedGlobal{
		conn: conn,
	}
}
