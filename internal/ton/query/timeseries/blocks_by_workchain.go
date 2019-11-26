package timeseries

import (
	"database/sql"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	selectBlocksCountByWorkchain = `SELECT 
	   groupArray(Time),
	   groupArray(BlocksMaster),
	   groupArray(BlocksWorkchain0)
    FROM (
		SELECT 
			toUInt64(Time) as Time, 
			sumIf(Blocks, WorkchainId = -1) as BlocksMaster,
			sumIf(Blocks, WorkchainId = 0) as BlocksWorkchain0
		FROM _ts_BlocksByWorkchain 
		WHERE Time <= now() AND Time >= now()-?
		GROUP BY Time
		ORDER BY Time
	);
`
)

type BlocksByWorkchain struct {
	Time        []uint64 `json:"time"`
	Masterchain []uint64 `json:"masterchain"`
	Workchain0  []uint64 `json:"workchain0"`
}

type GetBlocksByWorkchain struct {
	conn        *sql.DB
	resultCache *cache.WithTimer
}

func (q *GetBlocksByWorkchain) GetBlocksByWorkchain() (*BlocksByWorkchain, error) {
	if res, ok := q.resultCache.Get(); ok {
		switch res.(type) {
		case *BlocksByWorkchain:
			return res.(*BlocksByWorkchain), nil
		}
	}

	row := q.conn.QueryRow(selectBlocksCountByWorkchain, []byte("INTERVAL 8 MINUTE"))

	var resp = &BlocksByWorkchain{
		Time:        make([]uint64, 8*60/5), // 8 min by 5 sec per point
		Masterchain: make([]uint64, 8*60/5),
		Workchain0:  make([]uint64, 8*60/5),
	}

	if err := row.Scan(&resp.Time, &resp.Masterchain, &resp.Workchain0); err != nil {
		return nil, err
	}

	q.resultCache.Set(resp)

	return resp, nil
}

func NewGetBlocksByWorkchain(conn *sql.DB) *GetBlocksByWorkchain {
	return &GetBlocksByWorkchain{
		conn:        conn,
		resultCache: cache.NewWithTimer(time.Second),
	}
}
