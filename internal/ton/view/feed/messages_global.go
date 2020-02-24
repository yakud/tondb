package feed

import (
	"database/sql"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	createMessagesFeedGlobal = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_MessagesFeedGlobal
	ENGINE = MergeTree() 
	PARTITION BY toYYYYMM(Time)
	ORDER BY (Time, WorkchainId, Lt, MessageLt)
	SETTINGS index_granularity=128,index_granularity_bytes=0
	POPULATE 
	AS
	SELECT
		WorkchainId,
		Shard,
		SeqNo,
		Lt,
		Time,
	    Messages.CreatedLt as MessageLt, 
	    Messages.Direction as Direction, 
		Messages.SrcWorkchainId AS SrcWorkchainId, 
		Messages.SrcAddr AS Src, 
		Messages.DestWorkchainId AS DestWorkchainId, 
		Messages.DestAddr AS Dest, 
		Messages.ValueNanograms as ValueNanograms,
	    Messages.FwdFeeNanograms + Messages.IhrFeeNanograms + Messages.ImportFeeNanograms as TotalFeeNanograms, 
		Messages.Bounce as Bounce
	FROM transactions
	ARRAY JOIN Messages
	WHERE Type = 'trans_ord' AND Messages.Type = 'int_msg_info'
`
	dropMessagesFeedGlobal = `DROP TABLE _view_feed_MessagesFeedGlobal`

	querySelectMessages = `
	WITH (
		SELECT (min(Time), max(Time))
		FROM (
			SELECT 
			   Time
			FROM ".inner._view_feed_MessagesFeedGlobal"
			PREWHERE
				 if(? != 0, Time < toDateTime(?), 1) AND
				 if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId = ?, 1)
			ORDER BY Time DESC, WorkchainId DESC, Lt DESC, MessageLt DESC
			LIMIT ?
		)
	) as TimeRange
	SELECT 
		WorkchainId,
		hex(Shard),
		SeqNo,
		Lt,
		toUInt64(Time),
		Direction,
		SrcWorkchainId,
		Src,
		DestWorkchainId,
		Dest,
		ValueNanograms,
		TotalFeeNanograms,
		Bounce
	FROM ".inner._view_feed_MessagesFeedGlobal"
	PREWHERE 
		 (Time >= TimeRange.1 AND Time <= TimeRange.2) AND
		 if(? != bitShiftLeft(toInt32(-1), 31), WorkchainId = ?, 1)
	ORDER BY Time DESC, WorkchainId DESC, Lt DESC, MessageLt DESC
`
)

type MessageInFeed struct {
	WorkchainId      int32  `json:"workchain_id"`
	Shard            string `json:"shard"`
	SeqNo            uint64 `json:"seq_no"`
	Lt               uint64 `json:"lt"`
	Time             uint64 `json:"time"`
	Direction        string `json:"direction"`
	SrcWorkchainId   int32  `json:"src_workchain_id"`
	Src              string `json:"src"`
	SrcUf            string `json:"src_uf"`
	DestWorkchainId  int32  `json:"dest_workchain_id"`
	Dest             string `json:"dest"`
	DestUf           string `json:"dest_uf"`
	ValueNanogram    uint64 `json:"value_nanogram"`
	TotalFeeNanogram uint64 `json:"total_fee_nanogram"`
	Bounce           bool   `json:"bounce"`
}

type MessagesFeed struct {
	view.View
	conn *sql.DB
}

func (t *MessagesFeed) CreateTable() error {
	_, err := t.conn.Exec(createMessagesFeedGlobal)
	return err
}

func (t *MessagesFeed) DropTable() error {
	_, err := t.conn.Exec(dropMessagesFeedGlobal)
	return err
}

func (t *MessagesFeed) SelectLatestMessages(wcId int32, limit int16, beforeTime time.Time) ([]*MessageInFeed, error) {
	beforeTimeInt := beforeTime.Unix()
	rows, err := t.conn.Query(
		querySelectMessages,
		beforeTimeInt, beforeTimeInt, wcId, wcId, limit, wcId, wcId,
	)
	if err != nil {
		if rows != nil {
			rows.Close()
		}

		return nil, err
	}

	res := make([]*MessageInFeed, 0, limit)
	for rows.Next() {
		row := &MessageInFeed{}
		err := rows.Scan(
			&row.WorkchainId,
			&row.Shard,
			&row.SeqNo,
			&row.Lt,
			&row.Time,
			&row.Direction,
			&row.SrcWorkchainId,
			&row.Src,
			&row.DestWorkchainId,
			&row.Dest,
			&row.ValueNanogram,
			&row.TotalFeeNanogram,
			&row.Bounce,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}
		if row.SrcUf, err = utils.ComposeRawAndConvertToUserFriendly(row.SrcWorkchainId, row.Src); err != nil {
			// Maybe we shouldn't fail here?
			return nil, err
		}
		if row.DestUf, err = utils.ComposeRawAndConvertToUserFriendly(row.DestWorkchainId, row.Dest); err != nil {
			// Maybe we shouldn't fail here?
			return nil, err
		}

		res = append(res, row)
	}

	if rows != nil {
		rows.Close()
	}

	return res, err
}

func NewMessagesFeed(conn *sql.DB) *MessagesFeed {
	return &MessagesFeed{
		conn: conn,
	}
}
