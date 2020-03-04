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
		toUInt64(Time) AS TimeUnix, 
		Hash, 
		Messages.CreatedLt AS MessageLt, 
		Messages.Direction AS Direction, 
		Messages.SrcWorkchainId AS SrcWorkchainId, 
		Messages.SrcAddr AS Src, 
		Messages.DestWorkchainId AS DestWorkchainId, 
		Messages.DestAddr AS Dest, 
		Messages.ValueNanograms AS ValueNanograms, 
		Messages.FwdFeeNanograms + Messages.IhrFeeNanograms + Messages.ImportFeeNanograms AS TotalFeeNanograms, 
		Messages.Bounce AS Bounce, 
		Messages.BodyValue AS BodyValue
	FROM(
 		SELECT 
  			WorkchainId, 
    		Shard, 
    		SeqNo, 
    		Lt, 
    		Time, 
    		Hash, 
    		Messages.CreatedLt, 
    		Messages.Direction, 
    		Messages.SrcWorkchainId, 
    		Messages.SrcAddr, 
    		Messages.DestWorkchainId, 
    		Messages.DestAddr, 
    		Messages.ValueNanograms, 
    		Messages.FwdFeeNanograms,
    		Messages.IhrFeeNanograms,
 		    Messages.ImportFeeNanograms, 
    		Messages.Bounce, 
    		Messages.BodyValue
 		FROM transactions
 		PREWHERE ((WorkchainId, Shard, SeqNo) IN (
  			SELECT 
   				WorkchainId, 
   				Shard, 
   				SeqNo
  			FROM ".inner._view_index_TransactionBlock"
   			PREWHERE cityHash64(?) = Hash
 		)) AND (Hash = ?)
	) ARRAY JOIN Messages 
	WHERE (MessageLt = ?)
`
)

type GetMessage struct {
	conn *sql.DB
}

func (t *GetMessage) SelectMessage(trxHash string, messageLt uint64) (*feed.MessageInFeed, error) {
	msg := &feed.MessageInFeed{}
	row := t.conn.QueryRow(querySelectMessage, trxHash, trxHash, messageLt)
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