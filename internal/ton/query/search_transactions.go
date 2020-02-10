package query

import (
	"database/sql"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/utils"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

const (
	// todo: very ugly query, beautify it later
	querySelectTransactionsByFilter = `
	SELECT
		WorkchainId,
		Shard,
		SeqNo,
		Hash,
		Type,
		Lt,
		Time,
		TotalFeesNanograms,
		TotalFeesNanogramsLen,
		AccountAddr,
		OrigStatus,
		EndStatus,
		PrevTransLt,
		PrevTransHash,
		StateUpdateNewHash,
		StateUpdateOldHash,
		Messages.Direction as MessagesDirection,
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
	FROM transactions
	PREWHERE %s
	LIMIT 1000
`
)

type SearchTransactions struct {
	conn *sql.DB
}

func (s *SearchTransactions) SearchByFilter(f filter.Filter) ([]*ton.Transaction, error) {
	query, args, err := filter.RenderQuery(querySelectTransactionsByFilter, f)
	if err != nil {
		return nil, err
	}

	rows, err := s.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}

	transactions := make([]*ton.Transaction, 0)
	for rows.Next() {
		transaction := &ton.Transaction{
			OutMsgs: make([]*ton.TransactionMessage, 0),
		}
		messagesDirection := make([]string, 0)
		messagesType := make([]string, 0)
		messagesInit := make([]string, 0)
		messagesBounce := make([]uint8, 0)
		messagesBounced := make([]uint8, 0)
		messagesCreatedAt := make([]uint64, 0)
		messagesCreatedLt := make([]uint64, 0)
		messagesValueNanograms := make([]uint64, 0)
		messagesValueNanogramsLen := make([]uint8, 0)
		messagesFwdFeeNanograms := make([]uint64, 0)
		messagesFwdFeeNanogramsLen := make([]uint8, 0)
		messagesIhrDisabled := make([]uint8, 0)
		messagesIhrFeeNanograms := make([]uint64, 0)
		messagesIhrFeeNanogramsLen := make([]uint8, 0)
		messagesImportFeeNanograms := make([]uint64, 0)
		messagesImportFeeNanogramsLen := make([]uint8, 0)
		messagesDestIsEmpty := make([]uint8, 0)
		messagesDestWorkchainId := make([]int32, 0)
		messagesDestAddr := make([]string, 0)
		messagesDestAnycast := make([]string, 0)
		messagesSrcIsEmpty := make([]uint8, 0)
		messagesSrcWorkchainId := make([]int32, 0)
		messagesSrcAddr := make([]string, 0)
		messagesSrcAnycast := make([]string, 0)
		messagesBodyType := make([]string, 0)
		messagesBodyValue := make([]string, 0)
		trTime := &time.Time{}
		err = rows.Scan(
			&transaction.WorkchainId,
			&transaction.Shard,
			&transaction.SeqNo,
			&transaction.Hash,
			&transaction.Type,
			&transaction.Lt,
			&trTime,
			&transaction.TotalFeesNanograms,
			&transaction.TotalFeesNanogramsLen,
			&transaction.AccountAddr,
			&transaction.OrigStatus,
			&transaction.EndStatus,
			&transaction.PrevTransLt,
			&transaction.PrevTransHash,
			&transaction.StateUpdateNewHash,
			&transaction.StateUpdateOldHash,
			&messagesDirection,
			&messagesType,
			&messagesInit,
			&messagesBounce,
			&messagesBounced,
			&messagesCreatedAt,
			&messagesCreatedLt,
			&messagesValueNanograms,
			&messagesValueNanogramsLen,
			&messagesFwdFeeNanograms,
			&messagesFwdFeeNanogramsLen,
			&messagesIhrDisabled,
			&messagesIhrFeeNanograms,
			&messagesIhrFeeNanogramsLen,
			&messagesImportFeeNanograms,
			&messagesImportFeeNanogramsLen,
			&messagesDestIsEmpty,
			&messagesDestWorkchainId,
			&messagesDestAddr,
			&messagesDestAnycast,
			&messagesSrcIsEmpty,
			&messagesSrcWorkchainId,
			&messagesSrcAddr,
			&messagesSrcAnycast,
			&messagesBodyType,
			&messagesBodyValue,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}

		transaction.AccountAddrUf, err = utils.ConvertRawToUserFriendly(
			fmt.Sprintf("%d:%s", transaction.WorkchainId, transaction.AccountAddr), utils.DefaultTag)
		if err != nil {
			return nil, err
		}

		transaction.Now = uint64(trTime.Unix())
		for i, _ := range messagesDirection {
			direction := messagesDirection[i]
			var srcUf, destUf string

			if messagesDestIsEmpty[i] != 1 {
				destUf, err = utils.ConvertRawToUserFriendly(
					fmt.Sprintf("%d:%s", messagesDestWorkchainId[i], messagesDestAddr[i]), utils.DefaultTag)
				if err != nil {
					// Maybe we shouldn't fail here?
					return nil, err
				}
			}

			if messagesSrcIsEmpty[i] != 1 {
				srcUf, err = utils.ConvertRawToUserFriendly(
					fmt.Sprintf("%d:%s", messagesSrcWorkchainId[i], messagesSrcAddr[i]), utils.DefaultTag)
				if err != nil {
					// Maybe we shouldn't fail here?
					return nil, err
				}
			}

			msg := &ton.TransactionMessage{
				Type:                  messagesType[i],
				Init:                  messagesInit[i],
				Bounce:                messagesBounce[i] == 1,
				Bounced:               messagesBounced[i] == 1,
				CreatedAt:             messagesCreatedAt[i],
				CreatedLt:             messagesCreatedLt[i],
				ValueNanograms:        messagesValueNanograms[i],
				ValueNanogramsLen:     messagesValueNanogramsLen[i],
				FwdFeeNanograms:       messagesFwdFeeNanograms[i],
				FwdFeeNanogramsLen:    messagesFwdFeeNanogramsLen[i],
				IhrDisabled:           messagesIhrDisabled[i] == 1,
				IhrFeeNanograms:       messagesIhrFeeNanograms[i],
				IhrFeeNanogramsLen:    messagesIhrFeeNanogramsLen[i],
				ImportFeeNanograms:    messagesImportFeeNanograms[i],
				ImportFeeNanogramsLen: messagesImportFeeNanogramsLen[i],
				Dest: ton.AddrStd{
					IsEmpty:     messagesDestIsEmpty[i] == 1,
					WorkchainId: messagesDestWorkchainId[i],
					Addr:        messagesDestAddr[i],
					AddrUf:      destUf,
					Anycast:     messagesDestAnycast[i],
				},
				Src: ton.AddrStd{
					IsEmpty:     messagesSrcIsEmpty[i] == 1,
					WorkchainId: messagesSrcWorkchainId[i],
					Addr:        messagesSrcAddr[i],
					AddrUf:      srcUf,
					Anycast:     messagesSrcAnycast[i],
				},
				BodyType:  messagesBodyType[i],
				BodyValue: messagesBodyValue[i],
			}
			if direction == "in" {
				transaction.InMsg = msg
			}
			if direction == "out" {
				transaction.OutMsgs = append(transaction.OutMsgs, msg)
			}
			if msg.Src.IsEmpty {
				msg.Src.Addr = ""
			}
			if msg.Dest.IsEmpty {
				msg.Dest.Addr = ""
			}
		}
		transactions = append(transactions, transaction)
	}
	rows.Close()

	return transactions, nil
}

func NewSearchTransactions(conn *sql.DB) *SearchTransactions {
	return &SearchTransactions{
		conn: conn,
	}
}
