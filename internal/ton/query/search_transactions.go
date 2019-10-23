package query

import (
	"database/sql"
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
		groupArray(MessageDirection),
		groupArray(MessageType),
		groupArray(MessageInit),
		groupArray(MessageBounce),
		groupArray(MessageBounced),
		groupArray(MessageCreatedAt),
		groupArray(MessageCreatedLt),
		groupArray(MessageValueNanograms),
		groupArray(MessageValueNanogramsLen),
		groupArray(MessageFwdFeeNanograms),
		groupArray(MessageFwdFeeNanogramsLen),
		groupArray(MessageIhrDisabled),
		groupArray(MessageIhrFeeNanograms),
		groupArray(MessageIhrFeeNanogramsLen),
		groupArray(MessageImportFeeNanograms),
		groupArray(MessageImportFeeNanogramsLen),
		groupArray(MessageDestIsEmpty),
		groupArray(MessageDestWorkchainId),
		groupArray(MessageDestAddr),
		groupArray(MessageDestAnycast),
		groupArray(MessageSrcIsEmpty),
		groupArray(MessageSrcWorkchainId),
		groupArray(MessageSrcAddr),
		groupArray(MessageSrcAnycast),
		groupArray(MessageBodyType),
		groupArray(MessageBodyValue)
	FROM (
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
			Message.Direction as MessageDirection,
			Message.Type as MessageType,
			Message.Init as MessageInit,
			Message.Bounce as MessageBounce,
			Message.Bounced as MessageBounced,
			Message.CreatedAt as MessageCreatedAt,
			Message.CreatedLt as MessageCreatedLt,
			Message.ValueNanograms as MessageValueNanograms,
			Message.ValueNanogramsLen as MessageValueNanogramsLen,
			Message.FwdFeeNanograms as MessageFwdFeeNanograms,
			Message.FwdFeeNanogramsLen as MessageFwdFeeNanogramsLen,
			Message.IhrDisabled as MessageIhrDisabled,
			Message.IhrFeeNanograms as MessageIhrFeeNanograms,
			Message.IhrFeeNanogramsLen as MessageIhrFeeNanogramsLen,
			Message.ImportFeeNanograms as MessageImportFeeNanograms,
			Message.ImportFeeNanogramsLen as MessageImportFeeNanogramsLen,
			Message.DestIsEmpty as MessageDestIsEmpty,
			Message.DestWorkchainId as MessageDestWorkchainId,
			Message.DestAddr as MessageDestAddr,
			Message.DestAnycast as MessageDestAnycast,
			Message.SrcIsEmpty as MessageSrcIsEmpty,
			Message.SrcWorkchainId as MessageSrcWorkchainId,
			Message.SrcAddr as MessageSrcAddr,
			Message.SrcAnycast as MessageSrcAnycast,
			Message.BodyType as MessageBodyType,
			Message.BodyValue as MessageBodyValue
		FROM transactions
		ARRAY JOIN Messages as Message
		WHERE %s
	    LIMIT 1000
	) GROUP BY WorkchainId,Shard,SeqNo,Hash,Type,Lt,Time,TotalFeesNanograms,TotalFeesNanogramsLen,AccountAddr,OrigStatus,EndStatus,PrevTransLt,PrevTransHash,StateUpdateNewHash,StateUpdateOldHash
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

		transaction.Now = uint64(trTime.Unix())
		for i, _ := range messagesDirection {
			direction := messagesDirection[i]
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
					Anycast:     messagesDestAnycast[i],
				},
				Src: ton.AddrStd{
					IsEmpty:     messagesSrcIsEmpty[i] == 1,
					WorkchainId: messagesSrcWorkchainId[i],
					Addr:        messagesSrcAddr[i],
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
