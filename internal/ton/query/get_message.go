package query

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

const (
	querySelectMessage = `
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
	PREWHERE TrxHash = ? AND MessageLt = ?
`
)

type GetMessage struct {
	conn *sql.DB
}

func (t *GetMessage) SelectMessage(trxHash string, messageLt uint64) (*feed.MessageInFeed, error) {
	msg := &feed.MessageInFeed{}
	row := t.conn.QueryRow(querySelectMessage, trxHash, messageLt)
	if err := row.Scan(&msg.WorkchainId, &msg.Shard, &msg.SeqNo, &msg.Lt, &msg.Time, &msg.TrxHash, &msg.MessageLt,
		&msg.Direction, &msg.DestWorkchainId, &msg.Dest, &msg.SrcWorkchainId, &msg.Src, &msg.ValueNanogram,
		&msg.TotalFeeNanogram, &msg.Bounce, &msg.Body); err != nil {
		return nil, err
	}

	return msg, nil
}

func NewGetMessage(conn *sql.DB) *GetMessage {
	return &GetMessage{
		conn: conn,
	}
}