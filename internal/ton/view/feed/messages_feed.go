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
		toUInt64(Time) as TimeUnix,
		TrxHash,
		MessageLt,
		Direction,
		SrcWorkchainId,
		Src,
		DestWorkchainId,
		Dest,
		ValueNanograms,
		TotalFeeNanograms,
		Bounce,
	    BodyValue
	FROM ".inner._view_feed_MessagesFeedGlobal"
	PREWHERE 
		 (Time >= TimeRange.1 AND Time <= TimeRange.2) AND
		 (Lt <= TimeRange.3 AND MessageLt <= TimeRange.4)
	ORDER BY Time DESC, Lt DESC, MessageLt DESC, WorkchainId DESC
	LIMIT :limit
`
)

type MessageInFeed struct {
	WorkchainId      int32  `db:"WorkchainId" json:"workchain_id"`
	Shard            uint64 `db:"Shard" json:"shard"`
	SeqNo            uint64 `db:"SeqNo" json:"seq_no"`
	Lt               uint64 `db:"Lt" json:"lt"`
	Time             uint64 `db:"TimeUnix" json:"time"`
	TrxHash          string `db:"TrxHash" json:"trx_hash"`
	MessageLt        uint64 `db:"MessageLt" json:"message_lt"`
	Direction        string `db:"Direction" json:"direction"`
	SrcWorkchainId   int32  `db:"SrcWorkchainId" json:"src_workchain_id"`
	Src              string `db:"Src" json:"src"`
	SrcUf            string `db:"-" json:"src_uf"`
	DestWorkchainId  int32  `db:"DestWorkchainId" json:"dest_workchain_id"`
	Dest             string `db:"Dest" json:"dest"`
	DestUf           string `db:"-" json:"dest_uf"`
	ValueNanogram    uint64 `db:"ValueNanograms" json:"value_nanogram"`
	TotalFeeNanogram uint64 `db:"TotalFeeNanograms" json:"total_fee_nanogram"`
	Bounce           bool   `db:"Bounce" json:"bounce"`
	Body             string `db:"BodyValue" json:"body"`
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
