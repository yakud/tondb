package storage

import (
	"database/sql"
	"strings"
	"time"

	"github.com/mailru/go-clickhouse"

	"gitlab.flora.loc/mills/tondb/internal/utils"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

const (
	queryCreateTableTransactions string = `CREATE TABLE IF NOT EXISTS transactions (
		WorkchainId           Int32,
		Shard                 UInt64,
		SeqNo                 UInt64,
		Type                  LowCardinality(String),
		Lt                    UInt64,
		Time                  DateTime, -- field Now
		TotalFeesNanograms    UInt64,
		TotalFeesNanogramsLen UInt8,
		AccountAddr           FixedString(64),
		OrigStatus            LowCardinality(String),
		EndStatus             LowCardinality(String),
		PrevTransLt   		  UInt64,
		PrevTransHash 		  FixedString(64),
		StateUpdateNewHash    FixedString(64),
		StateUpdateOldHash    FixedString(64),
		
		Messages Nested
    	(
			Direction             LowCardinality(String),
			Type                  LowCardinality(String),
			Init                  LowCardinality(String),
			Bounce                UInt8,
			Bounced               UInt8,
			CreatedAt             UInt64,
			CreatedLt             UInt64,
			ValueNanograms        UInt64,
			ValueNanogramsLen     UInt8,
		    FwdFeeNanograms       UInt64,
			FwdFeeNanogramsLen    UInt8,
			IhrDisabled           UInt8,
			IhrFeeNanograms       UInt64,
			IhrFeeNanogramsLen    UInt8,
			ImportFeeNanograms    UInt64,
			ImportFeeNanogramsLen UInt8,
		    DestIsEmpty           UInt8,
			DestWorkchainId       Int32,
			DestAddr              FixedString(64),
			DestAnycast           LowCardinality(String),
		    SrcIsEmpty            UInt8,
			SrcWorkchainId        Int32,
			SrcAddr               FixedString(64),
			SrcAnycast            LowCardinality(String),
		    BodyType              LowCardinality(String),
			BodyValue             String
		)
	) ENGINE MergeTree
	PARTITION BY toYYYYMM(Time)
	ORDER BY (WorkchainId, Shard, SeqNo, Lt);
`

	queryInsertTransaction = `INSERT INTO transactions (
	WorkchainId,
	Shard,
	SeqNo,
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
	Messages.Direction,
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
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	queryDropTransactions = `DROP TABLE transactions;`
)

type Transactions struct {
	conn *sql.DB
}

func (s *Transactions) CreateTable() error {
	bdTx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	if _, err := bdTx.Exec(queryCreateTableTransactions); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Transactions) DropTable() error {
	bdTx, err := s.conn.Begin()
	if err != nil {
		return err
	}

	if _, err := bdTx.Exec(queryDropTransactions); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Transactions) InsertMany(transactions []*ton.Transaction) error {
	bdTx, err := s.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := s.InsertManyExec(transactions, bdTx)
	if err != nil {
		if stmt != nil {
			stmt.Close()
		}
		return err
	}

	if err := bdTx.Commit(); err != nil {
		if stmt != nil {
			stmt.Close()
		}
		return err
	}
	stmt.Close()

	return nil
}

func (s *Transactions) InsertManyExec(transactions []*ton.Transaction, bdTx *sql.Tx) (*sql.Stmt, error) {
	stmt, err := bdTx.Prepare(queryInsertTransaction)
	if err != nil {
		return stmt, err
	}

	for _, tr := range transactions {
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

		for _, msg := range tr.OutMsgs {
			messagesDirection = append(messagesDirection, "out")
			messagesType = append(messagesType, msg.Type)
			messagesInit = append(messagesInit, msg.Init)
			messagesBounce = append(messagesBounce, utils.BoolToUint8(msg.Bounce))
			messagesBounced = append(messagesBounced, utils.BoolToUint8(msg.Bounced))
			messagesCreatedAt = append(messagesCreatedAt, msg.CreatedAt)
			messagesCreatedLt = append(messagesCreatedLt, msg.CreatedLt)
			messagesValueNanograms = append(messagesValueNanograms, msg.ValueNanograms)
			messagesValueNanogramsLen = append(messagesValueNanogramsLen, msg.ValueNanogramsLen)
			messagesFwdFeeNanograms = append(messagesFwdFeeNanograms, msg.FwdFeeNanograms)
			messagesFwdFeeNanogramsLen = append(messagesFwdFeeNanogramsLen, msg.FwdFeeNanogramsLen)
			messagesIhrDisabled = append(messagesIhrDisabled, utils.BoolToUint8(msg.IhrDisabled))
			messagesIhrFeeNanograms = append(messagesIhrFeeNanograms, msg.IhrFeeNanograms)
			messagesIhrFeeNanogramsLen = append(messagesIhrFeeNanogramsLen, msg.IhrFeeNanogramsLen)
			messagesImportFeeNanograms = append(messagesImportFeeNanograms, msg.ImportFeeNanograms)
			messagesImportFeeNanogramsLen = append(messagesImportFeeNanogramsLen, msg.ImportFeeNanogramsLen)
			messagesDestIsEmpty = append(messagesDestIsEmpty, utils.BoolToUint8(msg.Dest.IsEmpty))
			messagesDestWorkchainId = append(messagesDestWorkchainId, msg.Dest.WorkchainId)
			messagesDestAddr = append(messagesDestAddr, strings.TrimLeft(msg.Dest.Addr, "x"))
			messagesDestAnycast = append(messagesDestAnycast, msg.Dest.Anycast)
			messagesSrcIsEmpty = append(messagesSrcIsEmpty, utils.BoolToUint8(msg.Src.IsEmpty))
			messagesSrcWorkchainId = append(messagesSrcWorkchainId, msg.Src.WorkchainId)
			messagesSrcAddr = append(messagesSrcAddr, strings.TrimLeft(msg.Src.Addr, "x"))
			messagesSrcAnycast = append(messagesSrcAnycast, msg.Src.Anycast)
			messagesBodyType = append(messagesBodyType, msg.BodyType)
			messagesBodyValue = append(messagesBodyValue, msg.BodyValue)
		}

		if tr.InMsg != nil {
			messagesDirection = append(messagesDirection, "in")
			messagesType = append(messagesType, tr.InMsg.Type)
			messagesInit = append(messagesInit, tr.InMsg.Init)
			messagesBounce = append(messagesBounce, utils.BoolToUint8(tr.InMsg.Bounce))
			messagesBounced = append(messagesBounced, utils.BoolToUint8(tr.InMsg.Bounced))
			messagesCreatedAt = append(messagesCreatedAt, tr.InMsg.CreatedAt)
			messagesCreatedLt = append(messagesCreatedLt, tr.InMsg.CreatedLt)
			messagesValueNanograms = append(messagesValueNanograms, tr.InMsg.ValueNanograms)
			messagesValueNanogramsLen = append(messagesValueNanogramsLen, tr.InMsg.ValueNanogramsLen)
			messagesFwdFeeNanograms = append(messagesFwdFeeNanograms, tr.InMsg.FwdFeeNanograms)
			messagesFwdFeeNanogramsLen = append(messagesFwdFeeNanogramsLen, tr.InMsg.FwdFeeNanogramsLen)
			messagesIhrDisabled = append(messagesIhrDisabled, utils.BoolToUint8(tr.InMsg.IhrDisabled))
			messagesIhrFeeNanograms = append(messagesIhrFeeNanograms, tr.InMsg.IhrFeeNanograms)
			messagesIhrFeeNanogramsLen = append(messagesIhrFeeNanogramsLen, tr.InMsg.IhrFeeNanogramsLen)
			messagesImportFeeNanograms = append(messagesImportFeeNanograms, tr.InMsg.ImportFeeNanograms)
			messagesImportFeeNanogramsLen = append(messagesImportFeeNanogramsLen, tr.InMsg.ImportFeeNanogramsLen)
			messagesDestIsEmpty = append(messagesDestIsEmpty, utils.BoolToUint8(tr.InMsg.Dest.IsEmpty))
			messagesDestWorkchainId = append(messagesDestWorkchainId, tr.InMsg.Dest.WorkchainId)
			messagesDestAddr = append(messagesDestAddr, strings.TrimLeft(tr.InMsg.Dest.Addr, "x"))
			messagesDestAnycast = append(messagesDestAnycast, tr.InMsg.Dest.Anycast)
			messagesSrcIsEmpty = append(messagesSrcIsEmpty, utils.BoolToUint8(tr.InMsg.Src.IsEmpty))
			messagesSrcWorkchainId = append(messagesSrcWorkchainId, tr.InMsg.Src.WorkchainId)
			messagesSrcAddr = append(messagesSrcAddr, strings.TrimLeft(tr.InMsg.Src.Addr, "x"))
			messagesSrcAnycast = append(messagesSrcAnycast, tr.InMsg.Src.Anycast)
			messagesBodyType = append(messagesBodyType, tr.InMsg.BodyType)
			messagesBodyValue = append(messagesBodyValue, tr.InMsg.BodyValue)
		}

		// in order like BlocksFields
		if _, err := stmt.Exec(
			tr.WorkchainId,
			tr.Shard,
			tr.SeqNo,
			tr.Type,
			tr.Lt,
			time.Unix(int64(tr.Now), 0).UTC(),
			tr.TotalFeesNanograms,
			tr.TotalFeesNanogramsLen,
			strings.TrimLeft(tr.AccountAddr, "x"),
			tr.OrigStatus,
			tr.EndStatus,
			tr.PrevTransLt,
			strings.TrimLeft(tr.PrevTransHash, "x"),
			strings.TrimLeft(tr.StateUpdateNewHash, "x"),
			strings.TrimLeft(tr.StateUpdateOldHash, "x"),

			clickhouse.Array(messagesDirection),
			clickhouse.Array(messagesType),
			clickhouse.Array(messagesInit),
			clickhouse.Array(messagesBounce),
			clickhouse.Array(messagesBounced),
			clickhouse.Array(messagesCreatedAt),
			clickhouse.Array(messagesCreatedLt),
			clickhouse.Array(messagesValueNanograms),
			clickhouse.Array(messagesValueNanogramsLen),
			clickhouse.Array(messagesFwdFeeNanograms),
			clickhouse.Array(messagesFwdFeeNanogramsLen),
			clickhouse.Array(messagesIhrDisabled),
			clickhouse.Array(messagesIhrFeeNanograms),
			clickhouse.Array(messagesIhrFeeNanogramsLen),
			clickhouse.Array(messagesImportFeeNanograms),
			clickhouse.Array(messagesImportFeeNanogramsLen),
			clickhouse.Array(messagesDestIsEmpty),
			clickhouse.Array(messagesDestWorkchainId),
			clickhouse.Array(messagesDestAddr),
			clickhouse.Array(messagesDestAnycast),
			clickhouse.Array(messagesSrcIsEmpty),
			clickhouse.Array(messagesSrcWorkchainId),
			clickhouse.Array(messagesSrcAddr),
			clickhouse.Array(messagesSrcAnycast),
			clickhouse.Array(messagesBodyType),
			clickhouse.Array(messagesBodyValue),
		); err != nil {
			return stmt, err
		}
	}

	return stmt, nil
}

func NewTransactions(conn *sql.DB) *Transactions {
	s := &Transactions{
		conn: conn,
	}

	return s
}
