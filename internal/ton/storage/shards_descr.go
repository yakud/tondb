package storage

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

const (
	queryCreateTableShardsDescr string = `CREATE TABLE IF NOT EXISTS shards_descr (
		MasterShard      UInt64,
		MasterSeqNo      UInt64,
		ShardWorkchainId Int32,
		Shard            UInt64,
		ShardSeqNo       UInt64
	)
	ENGINE MergeTree
	ORDER BY (MasterSeqNo, Shard, ShardSeqNo)
`

	queryInsertShardsDescr = `INSERT INTO shards_descr (MasterShard,MasterSeqNo,ShardWorkchainId,Shard,ShardSeqNo) VALUES (?,?,?,?,?);`
	queryDropShardsDescr   = `DROP TABLE shards_descr;`

	querySelectShardSeqRangesByMCSeq = `
	SELECT 
	   ? as MasterSeqNo,
	   ShardWorkchainId,
       Shard,
       min(ShardSeqNo)+1 as StartShardSeqno,
       max(ShardSeqNo) as EndShardSeqno
   	FROM (
		SELECT
		   MasterShard, 
		   MasterSeqNo, 
		   ShardWorkchainId,
		   Shard,
		   ShardSeqNo
		FROM shards_descr
		WHERE MasterSeqNo <= ? AND MasterSeqNo >= ?-50 AND Shard IN (SELECT Shard FROM shards_descr WHERE MasterSeqNo = ?)
		ORDER BY MasterSeqNo DESC, Shard DESC, ShardSeqNo DESC
		LIMIT 2 BY Shard
	) 
	GROUP BY MasterSeqNo,
	   ShardWorkchainId,
       Shard
`

	querySelectMCSeqByShardSeq = `
	SELECT 
	   MasterSeqNo
	FROM shards_descr
	WHERE Shard = ? AND ShardSeqNo >= ? AND ShardWorkchainId = ?
	ORDER BY MasterSeqNo ASC, Shard ASC, ShardSeqNo ASC
	LIMIT 1
`
)

type ShardsDescr struct {
	conn *sql.DB
}

func (c *ShardsDescr) CreateTable() error {
	bdTx, err := c.conn.Begin()
	if err != nil {
		return err
	}
	if _, err := bdTx.Exec(queryCreateTableShardsDescr); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *ShardsDescr) DropTable() error {
	bdTx, err := c.conn.Begin()
	if err != nil {
		return err
	}

	if _, err := bdTx.Exec(queryDropShardsDescr); err != nil {
		return err
	}

	if err := bdTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *ShardsDescr) InsertMany(rows []*ton.ShardDescr) error {
	bdTx, err := c.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := c.InsertManyExec(rows, bdTx)
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

func (c *ShardsDescr) InsertManyExec(rows []*ton.ShardDescr, bdTx *sql.Tx) (*sql.Stmt, error) {
	stmt, err := bdTx.Prepare(queryInsertShardsDescr)
	if err != nil {
		return stmt, err
	}

	for _, row := range rows {
		if _, err := stmt.Exec(
			row.MasterShard,
			row.MasterSeqNo,
			row.ShardWorkchainId,
			row.Shard,
			row.ShardSeqNo,
		); err != nil {
			return stmt, err
		}
	}

	return stmt, nil
}

type ShardBlocksRange struct {
	MasterSeq   uint64 `json:"master_seq"`
	WorkchainId int32  `json:"workchain_id"`
	Shard       uint64 `json:"shard"`
	FromSeq     uint64 `json:"from_seq"`
	ToSeq       uint64 `json:"to_seq"`
}

func (c *ShardsDescr) GetShardsSeqRangeInMasterBlock(masterSeq uint64) ([]ShardBlocksRange, error) {
	rows, err := c.conn.Query(querySelectShardSeqRangesByMCSeq, masterSeq, masterSeq, masterSeq, masterSeq)
	if err != nil {
		return nil, err
	}

	resp := make([]ShardBlocksRange, 0)
	for rows.Next() {
		s := ShardBlocksRange{}
		if err := rows.Scan(&s.MasterSeq, &s.WorkchainId, &s.Shard, &s.FromSeq, &s.ToSeq); err != nil {
			rows.Close()
			return nil, err
		}

		resp = append(resp, s)
	}

	rows.Close()

	return resp, nil
}

// todo: make it faster. sooo slow
func (c *ShardsDescr) GetMasterByShardBlock(shard *ton.BlockId) (*ton.BlockId, error) {
	rows, err := c.conn.Query(querySelectMCSeqByShardSeq, shard.Shard, shard.SeqNo, shard.WorkchainId)
	if err != nil {
		return nil, err
	}

	resp := &ton.BlockId{
		WorkchainId: -1,
		Shard:       0,
	}
	for rows.Next() {
		if err := rows.Scan(&resp.SeqNo); err != nil {
			rows.Close()
			return nil, err
		}
	}

	rows.Close()

	return resp, nil
}

func NewShardsDescr(conn *sql.DB) *ShardsDescr {
	s := &ShardsDescr{
		conn: conn,
	}

	return s
}
