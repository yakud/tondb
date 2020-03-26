package query

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
)

const (
	queryGetSyncedMasterHeightByLastShard = `
	WITH (
	    SELECT argMin((Shard, SeqNo), SeqNo) as t FROM (
			WITH (SELECT max(Time) FROM blocks) as maxTime
			SELECT 
				Shard,
				max(SeqNo) as SeqNo
			FROM blocks
			WHERE WorkchainId = 0 AND Time >= maxTime-INTERVAL 10 MINUTE AND Time < maxTime-5
			GROUP BY Shard
	   )) as LastSynced
	SELECT
	   MasterSeqNo
	FROM shards_descr
	WHERE Shard = LastSynced.1 AND ShardSeqNo >= LastSynced.2
	ORDER BY MasterSeqNo ASC, Shard ASC, ShardSeqNo ASC
	LIMIT 1
`
)

type GetSyncedHeight struct {
	conn *sql.DB
}

func (q *GetSyncedHeight) GetSyncedHeight() (*tonapi.BlockId, error) {
	row := q.conn.QueryRow(queryGetSyncedMasterHeightByLastShard)

	var syncedMasterHeight uint64
	if err := row.Scan(&syncedMasterHeight); err != nil {
		return nil, err
	}

	return &tonapi.BlockId{
		WorkchainId: -1,
		Shard:       0,
		SeqNo:       tonapi.Uint64(syncedMasterHeight),
	}, nil
}

func NewGetSyncedHeight(conn *sql.DB) *GetSyncedHeight {
	return &GetSyncedHeight{
		conn: conn,
	}
}
