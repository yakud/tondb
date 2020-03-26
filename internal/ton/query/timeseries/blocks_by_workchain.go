package timeseries

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
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

type GetBlocksByWorkchain struct {
	conn        *sql.DB
	resultCache *cache.WithTimer
}

func (q *GetBlocksByWorkchain) GetBlocksByWorkchain() (*tonapi.TimeseriesBlocksByWorkchain, error) {
	if res, ok := q.resultCache.Get(); ok {
		switch res.(type) {
		case *tonapi.TimeseriesBlocksByWorkchain:
			return res.(*tonapi.TimeseriesBlocksByWorkchain), nil
		}
	}

	row := q.conn.QueryRow(selectBlocksCountByWorkchain, []byte("INTERVAL 8 MINUTE"))

	times := make([]uint64, 8*60/5)
	masterchains := make([]uint64, 8*60/5)
	workchain0s := make([]uint64, 8*60/5)

	var resp = &tonapi.TimeseriesBlocksByWorkchain{
		Time:        make([]tonapi.Uint64, 0, 8*60/5), // 8 min by 5 sec per point
		Masterchain: make([]tonapi.Uint64, 0, 8*60/5),
		Workchain0:  make([]tonapi.Uint64, 0, 8*60/5),
	}

	if err := row.Scan(&times, &masterchains, &workchain0s); err != nil {
		return nil, err
	}

	for _, v := range times {
		resp.Time = append(resp.Time, tonapi.Uint64(v))
	}
	for _, v := range masterchains {
		resp.Masterchain= append(resp.Masterchain, tonapi.Uint64(v))
	}
	for _, v := range workchain0s {
		resp.Workchain0 = append(resp.Workchain0, tonapi.Uint64(v))
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
