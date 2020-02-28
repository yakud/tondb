package feed

import (
	"math"

	"gitlab.flora.loc/mills/tondb/internal/utils"

	"github.com/jmoiron/sqlx"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	EmptyWorkchainId     = math.MinInt32
	DefaultMessagesLimit = 50
	MaxMessagesLimit     = 500

	// TODO: change to ORDER BY (Time, Lt, MessageLt, WorkchainId)
	createMessagesFeedGlobal = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_MessagesFeedGlobal2
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
		Hash AS TrxHash,
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

	querySelectMessagesPart = `
	WITH (
		SELECT (min(Time), max(Time), max(Lt), max(MessageLt))
		FROM (
			SELECT 
			   Time,
			   Lt,
			   MessageLt
			FROM ".inner._view_feed_MessagesFeedGlobal"
			PREWHERE
				 if(:time == 0, 1,
					(Time = :time AND Lt <= :lt AND MessageLt < :message_lt) OR
					(Time < :time)
				 ) AND 
				 if(:workchain_id == bitShiftLeft(toInt32(-1), 31), 1, WorkchainId = :workchain_id)
			ORDER BY Time DESC, Lt DESC, MessageLt DESC, WorkchainId DESC
			LIMIT :limit
		)
	) as TimeRange
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		Lt,
		toUInt64(Time),
		TrxHash,
		MessageLt,
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
		 (Lt <= TimeRange.3 AND MessageLt <= TimeRange.4)
	ORDER BY Time DESC, Lt DESC, MessageLt DESC, WorkchainId DESC
	LIMIT :limit
`
)

type MessageInFeed struct {
	WorkchainId      int32  `json:"workchain_id"`
	Shard            uint64 `json:"shard"`
	SeqNo            uint64 `json:"seq_no"`
	Lt               uint64 `json:"lt"`
	Time             uint64 `json:"time"`
	TrxHash          string `json:"trx_hash"`
	MessageLt        uint64 `json:"message_lt"`
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

type MessagesFeedScrollId struct {
	Time        uint64 `json:"t"`
	Lt          uint64 `json:"l"`
	MessageLt   uint64 `json:"m"`
	WorkchainId int32  `json:"w"`
}

type messagesFeedDbFilter struct {
	Time        uint64 `db:"time"`
	Lt          uint64 `db:"lt"`
	MessageLt   uint64 `db:"message_lt"`
	Limit       uint16 `db:"limit"`
	WorkchainId int32  `db:"workchain_id"`
}

type MessagesFeed struct {
	view.View
	conn *sqlx.DB
}

func (t *MessagesFeed) CreateTable() error {
	_, err := t.conn.Exec(createMessagesFeedGlobal)
	return err
}

func (t *MessagesFeed) DropTable() error {
	_, err := t.conn.Exec(dropMessagesFeedGlobal)
	return err
}

func (t *MessagesFeed) SelectMessages(scrollId *MessagesFeedScrollId, limit uint16) ([]*MessageInFeed, *MessagesFeedScrollId, error) {
	if scrollId == nil {
		scrollId = &MessagesFeedScrollId{
			WorkchainId: EmptyWorkchainId,
		}
	}
	if limit == 0 {
		limit = DefaultMessagesLimit
	}
	if limit > MaxMessagesLimit {
		limit = MaxMessagesLimit
	}

	filter := messagesFeedDbFilter{
		Time:        scrollId.Time,
		Lt:          scrollId.Lt,
		MessageLt:   scrollId.MessageLt,
		Limit:       limit,
		WorkchainId: scrollId.WorkchainId,
	}

	rows, err := t.conn.NamedQuery(querySelectMessagesPart, &filter)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var feed []*MessageInFeed
	for rows.Next() {
		msg := &MessageInFeed{}
		if err := rows.StructScan(msg); err != nil {
			return nil, nil, err
		}
		if msg.SrcUf, err = utils.ComposeRawAndConvertToUserFriendly(msg.SrcWorkchainId, msg.Src); err != nil {
			// Maybe we shouldn't fail here?
			return nil, nil, err
		}
		if msg.DestUf, err = utils.ComposeRawAndConvertToUserFriendly(msg.DestWorkchainId, msg.Dest); err != nil {
			// Maybe we shouldn't fail here?
			return nil, nil, err
		}

		feed = append(feed, msg)
	}

	if len(feed) == 0 {
		return feed, nil, nil
	}
	var lastMsg = feed[len(feed)-1]
	newScrollId := &MessagesFeedScrollId{
		Time:        lastMsg.Time,
		Lt:          lastMsg.Lt,
		MessageLt:   lastMsg.MessageLt,
		WorkchainId: scrollId.WorkchainId,
	}

	return feed, newScrollId, nil
}

func NewMessagesFeed(conn *sqlx.DB) *MessagesFeed {
	return &MessagesFeed{
		conn: conn,
	}
}
