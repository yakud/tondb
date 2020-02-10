package feed

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/internal/utils"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	createFeedAccountTransactions = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_AccountTransactions
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
		Messages.Bounced as Bounced
	FROM transactions
	ARRAY JOIN Messages
`
	dropFeedAccountTransactions = `DROP TABLE _view_feed_AccountTransactions`

	querySelectAccountTransactions = `
	WITH (
		SELECT (min(Lt), max(Lt))
		FROM (
			SELECT Lt
			FROM ".inner._view_feed_AccountTransactions"
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
		Bounced
	FROM ".inner._view_feed_AccountTransactions"
	PREWHERE 
		(WorkchainId = ?) AND 
		(AccountAddr = ?) AND
		(Lt >= LtRange.1 AND Lt <= LtRange.2) AND
		%s
	ORDER BY WorkchainId DESC, AccountAddr DESC, Lt DESC, Time DESC
`
)

type AccountTransaction struct {
	WorkchainId        int32  `json:"workchain_id"`
	Shard              string `json:"shard"`
	SeqNo              uint64 `json:"seq_no"`
	AccountAddr        string `json:"account_addr"`
	AccountAddrUf      string `json:"account_addr_uf"`
	Lt                 uint64 `json:"lt"`
	Time               uint64 `json:"time"`
	Type               string `json:"type"`
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
}

type AccountTransactions struct {
	view.View
	conn *sql.DB
}

func (t *AccountTransactions) CreateTable() error {
	_, err := t.conn.Exec(createFeedAccountTransactions)
	return err
}

func (t *AccountTransactions) DropTable() error {
	_, err := t.conn.Exec(dropFeedAccountTransactions)
	return err
}

func (t *AccountTransactions) GetAccountTransactions(addr ton.AddrStd, afterLt uint64, count int16, f filter.Filter) ([]*AccountTransaction, error) {
	query, argsFilter, err := filter.RenderQuery(querySelectAccountTransactions, f)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		addr.WorkchainId, addr.Addr,
		afterLt, afterLt, count,
		addr.WorkchainId, addr.Addr,
	}
	args = append(args, argsFilter...)

	rows, err := t.conn.Query(query, args...)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return nil, err
	}

	res := make([]*AccountTransaction, 0, count)
	for rows.Next() {
		accTrans := &AccountTransaction{}
		err := rows.Scan(
			&accTrans.WorkchainId,
			&accTrans.Shard,
			&accTrans.SeqNo,
			&accTrans.AccountAddr,
			&accTrans.Lt,
			&accTrans.Time,
			&accTrans.Type,
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
		)

		accTrans.AccountAddrUf, err = utils.ConvertRawToUserFriendly(strconv.Itoa(int(accTrans.WorkchainId))+":"+accTrans.AccountAddr, utils.DefaultTag)
		if err != nil {
			// Maybe we shouldn't fail here?
			return nil, err
		}

		if accTrans.MessageType != "ext_in_msg_info" {
			accTrans.SrcUf, err = utils.ConvertRawToUserFriendly(strconv.Itoa(int(accTrans.SrcWorkchainId))+":"+accTrans.Src, utils.DefaultTag)
			if err != nil {
				// Maybe we shouldn't fail here?
				return nil, err
			}
		}

		accTrans.DestUf, err = utils.ConvertRawToUserFriendly(strconv.Itoa(int(accTrans.DestWorkchainId))+":"+accTrans.Dest, utils.DefaultTag)
		if err != nil {
			// Maybe we shouldn't fail here?
			return nil, err
		}

		if err != nil {
			rows.Close()
			return nil, err
		}

		res = append(res, accTrans)
	}

	rows.Close()

	return res, nil
}

func NewAccountTransactions(conn *sql.DB) *AccountTransactions {
	return &AccountTransactions{
		conn: conn,
	}
}
