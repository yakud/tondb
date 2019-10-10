package query

import (
	"database/sql"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton"
)

const (
	queryGetSyncedMasterHeightByLastShard = `
	WITH (
	    SELECT argMin((Shard, SeqNo), SeqNo) as t FROM (
			SELECT 
				toUInt64(bitOr(ShardPrefix, bitShiftLeft(1, (63 - toUInt64(ShardPfxBits))))) as Shard,
				max(SeqNo) as SeqNo
			FROM blocks
			WHERE ShardWorkchainId = 0
			GROUP BY Shard
	   )) as LastSynced
	SELECT
	   MasterSeqNo
	FROM shards_descr
	WHERE ShardPrefix = LastSynced.1 AND ShardSeqNo >= LastSynced.2
	ORDER BY MasterSeqNo ASC, ShardPrefix ASC, ShardSeqNo ASC
	LIMIT 1
`
)

type GetSyncedHeight struct {
	conn *sql.DB
}

func (q *GetSyncedHeight) GetSyncedHeight() (*ton.BlockId, error) {
	row := q.conn.QueryRow(queryGetSyncedMasterHeightByLastShard)

	var syncedMasterHeight uint64
	if err := row.Scan(&syncedMasterHeight); err != nil {
		return nil, err
	}

	return &ton.BlockId{
		WorkchainId: -1,
		ShardPrefix: 0,
		SeqNo:       syncedMasterHeight,
	}, nil
}

func NewGetSyncedHeight(conn *sql.DB) *GetSyncedHeight {
	return &GetSyncedHeight{
		conn: conn,
	}
}
