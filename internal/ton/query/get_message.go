package query

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	querySelectMessage = `
	SELECT  
		Messages.Type as MessagesType,
		Messages.Init as MessagesInit,
		Messages.Bounce as MessagesBounce,
		Messages.Bounced as MessagesBounced,
		Messages.CreatedAt as MessagesCreatedAt,
		Messages.CreatedLt as MessagesCreatedLt,
		Messages.ValueNanograms as MessagesValueNanograms,
		Messages.ValueNanogramsLen as MessagesValueNanogramsLen,
		Messages.FwdFeeNanograms as MessagesFwdFeeNanograms,
		Messages.FwdFeeNanogramsLen as MessagesFwdFeeNanogramsLen,
		Messages.IhrDisabled as MessagesIhrDisabled,
		Messages.IhrFeeNanograms as MessagesIhrFeeNanograms,
		Messages.IhrFeeNanogramsLen as MessagesIhrFeeNanogramsLen,
		Messages.ImportFeeNanograms as MessagesImportFeeNanograms,
		Messages.ImportFeeNanogramsLen as MessagesImportFeeNanogramsLen,
		Messages.DestIsEmpty as MessagesDestIsEmpty,
		Messages.DestWorkchainId as MessagesDestWorkchainId,
		Messages.DestAddr as MessagesDestAddr,
		Messages.DestAnycast as MessagesDestAnycast,
		Messages.SrcIsEmpty as MessagesSrcIsEmpty,
		Messages.SrcWorkchainId as MessagesSrcWorkchainId,
		Messages.SrcAddr as MessagesSrcAddr,
		Messages.SrcAnycast as MessagesSrcAnycast,
		Messages.BodyType as MessagesBodyType,
		Messages.BodyValue as MessagesBodyValue
	FROM(
 		SELECT 
			Messages.Type,
			Messages.Init,
			Messages.Bounce,
			Messages.Bounced,
			Messages.CreatedAt,
			Messages.CreatedLt,
			Messages.ValueNanograms,
			Messages.ValueNanogramsLen,
			Messages.FwdFeeNanograms,
			Messages.FwdFeeNanogramsLen,
			Messages.IhrDisabled,
			Messages.IhrFeeNanograms,
			Messages.IhrFeeNanogramsLen,
			Messages.ImportFeeNanograms,
			Messages.ImportFeeNanogramsLen,
			Messages.DestIsEmpty,
			Messages.DestWorkchainId,
			Messages.DestAddr,
			Messages.DestAnycast,
			Messages.SrcIsEmpty,
			Messages.SrcWorkchainId,
			Messages.SrcAddr,
			Messages.SrcAnycast,
			Messages.BodyType,
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
	WHERE (MessagesCreatedLt = ?)
`
)

type GetMessage struct {
	conn *sql.DB
}

func (t *GetMessage) SelectMessage(trxHash string, messageLt uint64) (msg *ton.TransactionMessage, err error) {
	msg = &ton.TransactionMessage{}
	src := ton.AddrStd{}
	dest := ton.AddrStd{}
	row := t.conn.QueryRow(querySelectMessage, trxHash, trxHash, messageLt)
	if err = row.Scan(&msg.Type, &msg.Init, &msg.Bounce, &msg.Bounced, &msg.CreatedAt, &msg.CreatedLt, &msg.ValueNanograms,
		&msg.ValueNanogramsLen, &msg.FwdFeeNanograms, &msg.FwdFeeNanogramsLen, &msg.IhrDisabled, &msg.IhrFeeNanograms,
		&msg.IhrFeeNanogramsLen, &msg.ImportFeeNanograms, &msg.ImportFeeNanogramsLen, &dest.IsEmpty, &dest.WorkchainId,
		&dest.Addr, &dest.Anycast, &src.IsEmpty, &src.WorkchainId, &src.Addr, &src.Anycast, &msg.BodyType, &msg.BodyValue); err != nil {
		return nil, err
	}

	if src.AddrUf, err = utils.ComposeRawAndConvertToUserFriendly(src.WorkchainId, src.Addr); err != nil {
		return nil, err
	}

	if dest.AddrUf, err = utils.ComposeRawAndConvertToUserFriendly(dest.WorkchainId, dest.Addr); err != nil {
		return nil, err
	}

	msg.Src = src
	msg.Dest = dest

	return msg, nil
}

func NewGetMessage(conn *sql.DB) *GetMessage {
	return &GetMessage{
		conn: conn,
	}
}