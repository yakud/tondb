package query

import (
	"database/sql"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/utils"

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
		ActionPhaseExists,
		ActionPhaseSuccess,
		ActionPhaseValid,
		ActionPhaseNoFunds,
		ActionPhaseCodeChanged,
		ActionPhaseActionListInvalid,
		ActionPhaseAccDeleteReq,
		ActionPhaseAccStatusChange,
		ActionPhaseTotalFwdFees,
		ActionPhaseTotalActionFees,
		ActionPhaseResultCode,
		ActionPhaseResultArg,
		ActionPhaseTotActions,
		ActionPhaseSpecActions,
		ActionPhaseSkippedActions,
		ActionPhaseMsgsCreated,
		ActionPhaseRemainingBalance,
		ActionPhaseReservedBalance,
		ActionPhaseEndLt,
		ActionPhaseTotMsgBits,
		ActionPhaseTotMsgCells,
		ComputePhaseExists,
		ComputePhaseSkipped,
		ComputePhaseSkippedReason,
		ComputePhaseAccountActivated,
		ComputePhaseSuccess,
		ComputePhaseMsgStateUsed,
		ComputePhaseOutOfGas,
		ComputePhaseAccepted,
		ComputePhaseExitArg,
		ComputePhaseExitCode,
		ComputePhaseMode,
		ComputePhaseVmSteps,
		ComputePhaseGasUsed,
		ComputePhaseGasMax,
		ComputePhaseGasCredit,
		ComputePhaseGasLimit,
		ComputePhaseGasFees,
		StoragePhaseExists,
		StoragePhaseStatus,
		StoragePhaseFeesCollected,
		StoragePhaseFeesDue,
		CreditPhaseExists,
		CreditPhaseDueFeesCollected,
		CreditPhaseCreditNanograms,
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
		arrayMap(body -> (
			if(
				(substr(body, 1, 10) = 'x{00000000' AND body != 'x{00000000}'),
				unhex(substring(replaceRegexpAll(body,'x{|}|\t|\n|\ ', ''), 9, length(body))),
	       		''
	   		 )
		), Messages.BodyValue) as MessagesBodyValue,
   		arraySum(Messages.ValueNanograms) as TotalNanograms,
	   	IsTock
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
		var isTock uint8
		actionPhase := &ton.ActionPhase{}
		computePhase := &ton.ComputePhase{}
		storagePhase := &ton.StoragePhase{}
		creditPhase := &ton.CreditPhase{}
		var actionPhaseExists, computePhaseExists, storagePhaseExists, creditPhaseExists bool

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
			&actionPhaseExists,
			&actionPhase.Success,
			&actionPhase.Valid,
			&actionPhase.NoFunds,
			&actionPhase.CodeChanged,
			&actionPhase.ActionListInvalid,
			&actionPhase.AccDeleteReq,
			&actionPhase.AccStatusChange,
			&actionPhase.TotalFwdFees,
			&actionPhase.TotalActionFees,
			&actionPhase.ResultCode,
			&actionPhase.ResultArg,
			&actionPhase.TotActions,
			&actionPhase.SpecActions,
			&actionPhase.SkippedActions,
			&actionPhase.MsgsCreated,
			&actionPhase.RemainingBalance,
			&actionPhase.ReservedBalance,
			&actionPhase.EndLt,
			&actionPhase.TotMsgBits,
			&actionPhase.TotMsgCells,
			&computePhaseExists,
			&computePhase.Skipped,
			&computePhase.SkippedReason,
			&computePhase.AccountActivated,
			&computePhase.Success,
			&computePhase.MsgStateUsed,
			&computePhase.OutOfGas,
			&computePhase.Accepted,
			&computePhase.ExitArg,
			&computePhase.ExitCode,
			&computePhase.Mode,
			&computePhase.VmSteps,
			&computePhase.GasUsed,
			&computePhase.GasMax,
			&computePhase.GasCredit,
			&computePhase.GasLimit,
			&computePhase.GasFees,
			&storagePhaseExists,
			&storagePhase.Status,
			&storagePhase.FeesCollected,
			&storagePhase.FeesDue,
			&creditPhaseExists,
			&creditPhase.DueFeesCollected,
			&creditPhase.CreditNanograms,
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
			&transaction.TotalNanograms,
			&isTock,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}

		if actionPhaseExists {
			transaction.ActionPhase = actionPhase
		} else {
			transaction.ActionPhase = nil
		}

		if computePhaseExists {
			transaction.ComputePhase = computePhase
		} else {
			transaction.ComputePhase = nil
		}

		if storagePhaseExists {
			transaction.StoragePhase = storagePhase
		} else {
			transaction.StoragePhase = nil
		}

		if creditPhaseExists {
			transaction.CreditPhase = creditPhase
		} else {
			transaction.CreditPhase = nil
		}

		transaction.AccountAddrUf, err = utils.ComposeRawAndConvertToUserFriendly(transaction.WorkchainId, transaction.AccountAddr)
		if err != nil {
			return nil, err
		}

		transaction.IsTock = isTock == 1

		transaction.Now = uint64(trTime.Unix())
		for i, _ := range messagesDirection {
			direction := messagesDirection[i]
			var srcUf, destUf string

			if messagesDestIsEmpty[i] != 1 {
				messagesDestAddr[i] = utils.NullAddrToString(messagesDestAddr[i])
				destUf, err = utils.ComposeRawAndConvertToUserFriendly(messagesDestWorkchainId[i], messagesDestAddr[i])
				if err != nil {
					// Maybe we shouldn't fail here?
					return nil, err
				}
			}

			if messagesSrcIsEmpty[i] != 1 {
				messagesSrcAddr[i] = utils.NullAddrToString(messagesSrcAddr[i])
				srcUf, err = utils.ComposeRawAndConvertToUserFriendly(messagesSrcWorkchainId[i], messagesSrcAddr[i])
				if err != nil {
					// Maybe we shouldn't fail here?
					return nil, err
				}
			}

			msg := &ton.TransactionMessage{
				TrxHash:               transaction.Hash,
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
