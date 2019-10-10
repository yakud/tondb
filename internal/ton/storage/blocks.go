package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton"
	"github.com/yakud/ton-blocks-stream-receiver/internal/utils"
)

type kv struct {
	k string
	v string
}

var blocksFields = []kv{
	{"ShardWorkchainId", "Int32"},
	{"ShardPrefix", "UInt64"},
	{"ShardPfxBits", "UInt8"},
	{"SeqNo", "UInt64"},
	{"Time", "DateTime"},

	{"MinRefMcSeqno", "UInt32"},
	{"PrevKeyBlockSeqno", "UInt32"},
	{"GenCatchainSeqno", "UInt32"},

	{"Prev1RefEndLt", "UInt64"},
	{"Prev1RefSeqNo", "UInt64"},
	{"Prev1RefFileHash", "FixedString(64)"},
	{"Prev1RefRootHash", "FixedString(64)"},

	{"Prev2RefEndLt", "UInt64"},
	{"Prev2RefSeqNo", "UInt64"},
	{"Prev2RefFileHash", "FixedString(64)"},
	{"Prev2RefRootHash", "FixedString(64)"},

	{"MasterRefEndLt", "UInt64"},
	{"MasterRefSeqNo", "UInt64"},
	{"MasterRefFileHash", "FixedString(64)"},
	{"MasterRefRootHash", "FixedString(64)"},

	{"StartLt", "UInt64"},
	{"EndLt", "UInt64"},
	{"Version", "UInt32"},
	{"Flags", "UInt8"},
	{"KeyBlock", "UInt8"},
	{"NotMaster", "UInt8"},
	{"WantMerge", "UInt8"},
	{"WantSplit", "UInt8"},
	{"AfterMerge", "UInt8"},
	{"AfterSplit", "UInt8"},
	{"BeforeSplit", "UInt8"},
}

const (
	queryCreateTableBloc string = `CREATE TABLE IF NOT EXISTS blocks (%s) 
	ENGINE MergeTree
	PARTITION BY toYYYYMM(Time)
	ORDER BY (ShardWorkchainId, ShardPrefix, SeqNo)
`

	queryInsertBlock = `INSERT INTO blocks (%s) VALUES (%s);`
	queryDropBlocks  = `DROP TABLE blocks;`
)

type Blocks struct {
	conn *sql.DB

	queryCreate string
	queryInsert string
}

func (c *Blocks) PrepareQueries() {
	kvCreate := make([]string, 0, len(blocksFields))
	kInsert := make([]string, 0, len(blocksFields))
	vInsert := make([]string, 0, len(blocksFields))

	for _, v := range blocksFields {
		kvCreate = append(kvCreate, fmt.Sprintf("%s %s", v.k, v.v))
		kInsert = append(kInsert, v.k)
		vInsert = append(vInsert, "?")
	}

	c.queryCreate = fmt.Sprintf(queryCreateTableBloc, strings.Join(kvCreate, ","))
	c.queryInsert = fmt.Sprintf(
		queryInsertBlock,
		strings.Join(kInsert, ","),
		strings.Join(vInsert, ","),
	)
}

func (c *Blocks) CreateTable() error {
	bdTx, err := c.conn.Begin()
	if err != nil {
		return err
	}
	if _, err := bdTx.Exec(c.queryCreate); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *Blocks) DropTable() error {
	bdTx, err := c.conn.Begin()
	if err != nil {
		return err
	}

	if _, err := bdTx.Exec(queryDropBlocks); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *Blocks) InsertMany(blocks []*ton.Block) error {
	bdTx, err := c.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := c.InsertManyExec(blocks, bdTx)
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

func (c *Blocks) InsertManyExec(rows []*ton.Block, bdTx *sql.Tx) (*sql.Stmt, error) {
	stmt, err := bdTx.Prepare(c.queryInsert)
	if err != nil {
		return stmt, err
	}

	for _, row := range rows {
		if row.Info.MasterRef == nil {
			row.Info.MasterRef = &ton.BlockRef{}
		}
		if row.Info.Prev1Ref == nil {
			row.Info.Prev1Ref = &ton.BlockRef{}
		}
		if row.Info.Prev2Ref == nil {
			row.Info.Prev2Ref = &ton.BlockRef{}
		}
		if _, err := stmt.Exec(
			row.Info.ShardWorkchainId,
			row.Info.ShardPrefix,
			row.Info.ShardPfxBits,
			row.Info.SeqNo,
			time.Unix(int64(row.Info.GenUtime), 0).UTC(),

			row.Info.MinRefMcSeqno,
			row.Info.PrevKeyBlockSeqno,
			row.Info.GenCatchainSeqno,

			row.Info.Prev1Ref.EndLt,
			row.Info.Prev1Ref.SeqNo,
			strings.TrimLeft(row.Info.Prev1Ref.FileHash, "x"),
			strings.TrimLeft(row.Info.Prev1Ref.RootHash, "x"),

			row.Info.Prev2Ref.EndLt,
			row.Info.Prev2Ref.SeqNo,
			strings.TrimLeft(row.Info.Prev2Ref.FileHash, "x"),
			strings.TrimLeft(row.Info.Prev2Ref.RootHash, "x"),

			row.Info.MasterRef.EndLt,
			row.Info.MasterRef.SeqNo,
			strings.TrimLeft(row.Info.MasterRef.FileHash, "x"),
			strings.TrimLeft(row.Info.MasterRef.RootHash, "x"),

			row.Info.StartLt,
			row.Info.EndLt,
			row.Info.Version,
			row.Info.Flags,

			utils.BoolToUint8(row.Info.KeyBlock),
			utils.BoolToUint8(row.Info.NotMaster),
			utils.BoolToUint8(row.Info.WantMerge),
			utils.BoolToUint8(row.Info.WantSplit),
			utils.BoolToUint8(row.Info.AfterMerge),
			utils.BoolToUint8(row.Info.AfterSplit),
			utils.BoolToUint8(row.Info.BeforeSplit),
		); err != nil {
			return stmt, err
		}
	}

	return stmt, nil
}

func NewBlocks(conn *sql.DB) *Blocks {
	s := &Blocks{
		conn: conn,
	}
	s.PrepareQueries()

	return s
}
