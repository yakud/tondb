package storage

import (
	"database/sql"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton"
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
       Shard,
       min(MinShardSeqNo)+1 as StartShardSeqno,
       max(MinShardSeqNo) as EndShardSeqno
   	FROM(
		SELECT
		   MasterShard, 
		   MasterSeqNo, 
		   Shard,
		   max(ShardSeqNo) as MinShardSeqNo
		FROM shards_descr
		WHERE MasterSeqNo = ?
		GROUP BY MasterShard, MasterSeqNo, Shard
		ORDER BY MasterShard DESC, MasterSeqNo DESC, Shard DESC
		
		UNION ALL 
		
		SELECT
		   MasterShard, 
		   MasterSeqNo, 
		   Shard,
		   max(ShardSeqNo) as MinShardSeqNo
		FROM shards_descr
		WHERE MasterSeqNo < ? AND Shard IN (
			SELECT
			   Shard
			FROM shards_descr
			WHERE MasterSeqNo = ?
			GROUP BY Shard
		)
		GROUP BY MasterShard, MasterSeqNo, Shard
		ORDER BY MasterShard DESC, MasterSeqNo DESC, Shard DESC
		LIMIT 1 BY Shard
	)
	GROUP BY Shard
	ORDER BY Shard
`

	querySelectMCSeqByShardSeq = `
	SELECT 
	   MasterSeqNo
	FROM shards_descr
	WHERE Shard = ? AND ShardSeqNo >= ?
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
	WorkchainId int32  `json:"workchain_id"`
	Shard       uint64 `json:"shard"`
	FromSeq     uint64 `json:"from_seq"`
	ToSeq       uint64 `json:"to_seq"`
}

func (c *ShardsDescr) GetShardsSeqRangeInMCBlock(mcSeqNum uint64) ([]ShardBlocksRange, error) {
	rows, err := c.conn.Query(querySelectShardSeqRangesByMCSeq, mcSeqNum, mcSeqNum, mcSeqNum)
	if err != nil {
		return nil, err
	}

	resp := make([]ShardBlocksRange, 0)
	for rows.Next() {
		s := ShardBlocksRange{
			WorkchainId: 0,
		}
		if err := rows.Scan(&s.Shard, &s.FromSeq, &s.ToSeq); err != nil {
			rows.Close()
			return nil, err
		}

		resp = append(resp, s)
	}

	rows.Close()

	return resp, nil
}

func (c *ShardsDescr) GetMCSeqByShardSeq(shard, shardSeq uint64) (*ton.BlockId, error) {
	rows, err := c.conn.Query(querySelectMCSeqByShardSeq, shard, shardSeq)
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
