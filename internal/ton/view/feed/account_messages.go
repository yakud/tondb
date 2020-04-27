package feed

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton/view"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	createFeedAccountMessages = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_AccountMessages
	ENGINE = MergeTree() 
	PARTITION BY toYYYYMM(Time)
	ORDER BY (WorkchainId, AccountAddr, Lt, Time)
	SETTINGS index_granularity=128,index_granularity_bytes=0
	POPULATE 
	AS
	SELECT
		WorkchainId,
		Shard,
		SeqNo,
		AccountAddr,
		Lt,
		Time,
		Type,
		Hash as TrxHash,
		Messages.Type as MessageType, 
		Messages.CreatedLt as MessageLt, 
	    Messages.Direction as Direction, 
		Messages.SrcWorkchainId AS SrcWorkchainId, 
		Messages.SrcAddr AS Src, 
		Messages.DestWorkchainId AS DestWorkchainId, 
		Messages.DestAddr AS Dest, 
		Messages.ValueNanograms as ValueNanograms,
	    Messages.FwdFeeNanograms as FwdFeeNanograms, 
		Messages.IhrFeeNanograms as IhrFeeNanograms,
		Messages.ImportFeeNanograms as ImportFeeNanograms,
		Messages.Bounce as Bounce,
		Messages.Bounced as Bounced,
		if(
	        (substr(Messages.BodyValue, 1, 10) = 'x{00000000' AND Messages.BodyValue != 'x{00000000}'),
	        unhex(substring(replaceRegexpAll(Messages.BodyValue,'x{|}|\t|\n|\ ', ''), 9, length(Messages.BodyValue))),
	        ''
	    ) AS BodyValue
	FROM transactions
	ARRAY JOIN Messages
`
	dropFeedAccountMessages = `DROP TABLE _view_feed_AccountMessages`

	querySelectAccountMessages = `
	WITH (
		SELECT (min(Lt), max(Lt))
		FROM (
			SELECT Lt
			FROM ".inner._view_feed_AccountMessages"
			PREWHERE 
				(WorkchainId = ?) AND 
				(AccountAddr = ?) AND
				if(?!=0, Lt < ?, 1 == 1)
			ORDER BY WorkchainId DESC, AccountAddr DESC, Lt DESC, Time DESC
			LIMIT ?
		)
	) as LtRange
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		AccountAddr,
		Lt,
		toUInt64(Time),
		Type,
		TrxHash,
		MessageType,
		MessageLt,
		Direction,
		SrcWorkchainId,
		Src,
		DestWorkchainId,
		Dest,
		ValueNanograms,
		FwdFeeNanograms,
		IhrFeeNanograms,
		ImportFeeNanograms,
		Bounce,
		Bounced,
		BodyValue
	FROM ".inner._view_feed_AccountMessages"
	PREWHERE 
		(WorkchainId = ?) AND 
		(AccountAddr = ?) AND
		(Lt >= LtRange.1 AND Lt <= LtRange.2) AND
		%s
	ORDER BY WorkchainId DESC, AccountAddr DESC, Lt DESC, Time DESC
`
)

type AccountMessage struct {
	WorkchainId        int32  `json:"workchain_id"`
	Shard              string `json:"shard"`
	SeqNo              uint64 `json:"seq_no"`
	AccountAddr        string `json:"account_addr"`
	AccountAddrUf      string `json:"account_addr_uf"`
	Lt                 uint64 `json:"lt"`
	Time               uint64 `json:"time"`
	Type               string `json:"type"`
	TrxHash            string `json:"trx_hash"`
	MessageType        string `json:"message_type"`
	MessageLt          uint64 `json:"message_lt"`
	Direction          string `json:"direction"`
	SrcWorkchainId     int32  `json:"src_workchain_id"`
	Src                string `json:"src"`
	SrcUf              string `json:"src_uf"`
	DestWorkchainId    int32  `json:"dest_workchain_id"`
	Dest               string `json:"dest"`
	DestUf             string `json:"dest_uf"`
	ValueNanograms     string `json:"value_nanograms"`
	FwdFeeNanograms    string `json:"fwd_fee_nanograms"`
	IhrFeeNanograms    string `json:"ihr_fee_nanograms"`
	ImportFeeNanograms string `json:"import_fee_nanograms"`
	Bounce             uint8  `json:"bounce"`
	Bounced            uint8  `json:"bounced"`
	Body               string `json:"body"`
}

type AccountMessages struct {
	view.View
	conn *sql.DB
}

type AccountMessagesScrollId struct {
	Lt uint64 `json:"lt"`
}

func (t *AccountMessages) CreateTable() error {
	_, err := t.conn.Exec(createFeedAccountMessages)
	return err
}

func (t *AccountMessages) DropTable() error {
	_, err := t.conn.Exec(dropFeedAccountMessages)
	return err
}

func (t *AccountMessages) GetAccountMessages(addr ton.AddrStd, scrollId *AccountMessagesScrollId, count int16, f filter.Filter) ([]*AccountMessage, *AccountMessagesScrollId, error) {
	query, argsFilter, err := filter.RenderQuery(querySelectAccountMessages, f)
	if err != nil {
		return nil, nil, err
	}

	args := []interface{}{
		addr.WorkchainId, addr.Addr,
		scrollId.Lt, scrollId.Lt, count,
		addr.WorkchainId, addr.Addr,
	}
	args = append(args, argsFilter...)

	rows, err := t.conn.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	res := make([]*AccountMessage, 0, count)
	for rows.Next() {
		accTrans := &AccountMessage{}
		err := rows.Scan(
			&accTrans.WorkchainId,
			&accTrans.Shard,
			&accTrans.SeqNo,
			&accTrans.AccountAddr,
			&accTrans.Lt,
			&accTrans.Time,
			&accTrans.Type,
			&accTrans.TrxHash,
			&accTrans.MessageType,
			&accTrans.MessageLt,
			&accTrans.Direction,
			&accTrans.SrcWorkchainId,
			&accTrans.Src,
			&accTrans.DestWorkchainId,
			&accTrans.Dest,
			&accTrans.ValueNanograms,
			&accTrans.FwdFeeNanograms,
			&accTrans.IhrFeeNanograms,
			&accTrans.ImportFeeNanograms,
			&accTrans.Bounce,
			&accTrans.Bounced,
			&accTrans.Body,
		)

		accTrans.AccountAddr = utils.NullAddrToString(accTrans.AccountAddr)
		accTrans.Src = utils.NullAddrToString(accTrans.Src)
		accTrans.Dest = utils.NullAddrToString(accTrans.Dest)

		accTrans.AccountAddrUf, err = utils.ComposeRawAndConvertToUserFriendly(accTrans.WorkchainId, accTrans.AccountAddr)
		if err != nil {
			// Maybe we shouldn't fail here?
			return nil, nil, err
		}

		if accTrans.MessageType != "ext_in_msg_info" {
			accTrans.SrcUf, err = utils.ComposeRawAndConvertToUserFriendly(accTrans.SrcWorkchainId, accTrans.Src)
			if err != nil {
				// Maybe we shouldn't fail here?
				return nil, nil, err
			}
		}

		accTrans.DestUf, err = utils.ComposeRawAndConvertToUserFriendly(accTrans.DestWorkchainId, accTrans.Dest)
		if err != nil {
			// Maybe we shouldn't fail here?
			return nil, nil, err
		}

		res = append(res, accTrans)
	}

	if len(res) == 0 {
		return res, nil, nil
	}

	newScrollId := &AccountMessagesScrollId{
		Lt: res[len(res)-1].Lt,
	}

	return res, newScrollId, nil
}

func NewAccountMessages(conn *sql.DB) *AccountMessages {
	return &AccountMessages{
		conn: conn,
	}
}
