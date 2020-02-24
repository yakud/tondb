package stats

import (
	"database/sql"
	"errors"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	getHeightAndTotalBlocks = `SELECT count() AS TotalBlocks, max(SeqNo) AS BlockHeight FROM ".inner._view_feed_BlocksFeed" %s`

	getAvgBlockTime = `
	SELECT                                                                                                              
   		avg(runningDifference(Time)) as AvgBlockTime
	FROM(            
 		SELECT
			Time 
 		FROM ".inner._view_feed_BlocksFeed" 
 		PREWHERE Time > now() - INTERVAL %d %s %s 
 		ORDER BY Time
	)
`

	// maybe I should move these consts somwhere else
	workchainIdPrewhere = "PREWHERE WorkchainId = %d"

	workchainIdAnd = "AND WorkchainId = %d"

	intervalDay = "DAY"

	intervalWeek = "WEEK"

	intervalMonth = "MONTH"

	intervalYear = "YEAR"

	cacheKeyBlocksMetrics = "blocks_metrics"
)

type BlocksMetricsResult struct {
	TotalBlocks  uint64 `json:"total_blocks"`
	BlocksHeight uint64 `json:"blocks_height"`
	AvgBlockTime float64 `json:"avg_block_time"`
}

type BlocksMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *BlocksMetrics) UpdateQuery() error {
	res := BlocksMetricsResult{}

	row := t.conn.QueryRow(fmt.Sprintf(getHeightAndTotalBlocks, ""))
	if err := row.Scan(&res.TotalBlocks, &res.BlocksHeight); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getAvgBlockTime, 1, intervalWeek, ""))
	if err := row.Scan(&res.AvgBlockTime); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyBlocksMetrics, &res)

	resWorkchain := BlocksMetricsResult{}

	row = t.conn.QueryRow(fmt.Sprintf(getHeightAndTotalBlocks, fmt.Sprintf(workchainIdPrewhere, 0)))
	if err := row.Scan(&resWorkchain.TotalBlocks, &resWorkchain.BlocksHeight); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getAvgBlockTime,  1, intervalWeek, fmt.Sprintf(workchainIdAnd, 0)))
	if err := row.Scan(&resWorkchain.AvgBlockTime); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyBlocksMetrics + "0", &resWorkchain)

	resMasterchain := BlocksMetricsResult{}

	row = t.conn.QueryRow(fmt.Sprintf(getHeightAndTotalBlocks, fmt.Sprintf(workchainIdPrewhere, -1)))
	if err := row.Scan(&resMasterchain.TotalBlocks, &resMasterchain.BlocksHeight); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getAvgBlockTime,  1, intervalWeek, fmt.Sprintf(workchainIdAnd, -1)))
	if err := row.Scan(&resMasterchain.AvgBlockTime); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyBlocksMetrics + "-1", &resMasterchain)

	return nil
}

func (t *BlocksMetrics) GetBlocksMetrics(workchainId string) (*BlocksMetricsResult, error) {
	if res, err := t.resultCache.Get(cacheKeyBlocksMetrics + workchainId); err == nil {
		switch res.(type) {
		case *BlocksMetricsResult:
			return res.(*BlocksMetricsResult), nil
		default:
			return nil, errors.New("couldn't get blocks metrics from cache, cache contains object of wrong type")
		}
	}

	return nil, errors.New("couldn't get blocks metrics from cache, wrong workchainId specified or cache is empty")
}

func NewBlocksMetrics(conn *sql.DB, cache cache.Cache) *BlocksMetrics {
	return &BlocksMetrics{
		conn:        conn,
		resultCache: cache,
	}
}

