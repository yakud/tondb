package stats

import (
	"database/sql"
	"errors"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	getGlobalMetrics = `
	SELECT
		sum(TotalAddr) AS TotalAddr,
	    sum(TotalNanogram) AS TotalNanogram,
	    sum(TotalMessages) AS TotalMessages,
	    sum(TotalBlocks) AS TotalBlocks,
	    sum(TrxLastDay) AS TrxLastDay,
	    sum(TrxLastMonth) AS TrxLastMonth
	FROM (
		SELECT
			count() AS TotalAddr,
			sum(BalanceNanogram) AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalBlocks,
		    0 AS TrxLastDay,
		    0 AS TrxLastMonth
		FROM ".inner._view_state_AccountState" FINAL

	 	UNION ALL
		
		SELECT 
		   	0 AS TotalAddr,
		   	0 AS TotalNanogram,
			sum(TotalMessages) AS TotalMessages,
			0 AS TotalBlocks,
			0 AS TrxLastDay,
		    0 AS TrxLastMonth
		FROM ".inner._view_feed_TotalTransactionsAndMessages"
		    
		UNION ALL  

		SELECT
		    0 AS TotalAddr,
		    0 AS TotalNanogram,
		    0 AS TotalMessages,
		    count() AS TotalBlocks,
    		sum(if(Time >= (now() - INTERVAL 1 DAY), BlockStatsTrxCount, 0)) AS TrxLastDay, 
    		sum(if(Time >= (now() - INTERVAL 1 MONTH), BlockStatsTrxCount, 0)) AS TrxLastMonth
		FROM blocks
	)
`

	cacheKeyGlobalMetrics = "global_metrics"
)

type GlobalMetricsResult struct {
	TotalAddr     uint64 `json:"total_addr"`
	TotalNanogram uint64 `json:"total_nanogram"`
	TotalBlocks   uint64 `json:"total_blocks"`
	TotalMessages uint64 `json:"total_messages"`
	TrxLastDay    uint64 `json:"trx_last_day"`
	TrxLastMonth  uint64 `json:"trx_last_month"`
}

type GlobalMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *GlobalMetrics) UpdateQuery() error {
	res := GlobalMetricsResult{}

	row := t.conn.QueryRow(getGlobalMetrics)

	if err := row.Scan(&res.TotalAddr, &res.TotalNanogram, &res.TotalMessages, &res.TotalBlocks, &res.TrxLastDay, &res.TrxLastMonth); err != nil {
		return err
	}

	// Not handling the error because it's quite useless in this implementation of Cache interface (cache.Background).
	// Set function returns error in Cache interface because in some implementations of it errors may be somehow valuable.
	t.resultCache.Set(cacheKeyGlobalMetrics, &res)

	return nil
}

func (t *GlobalMetrics) GetGlobalMetrics() (*GlobalMetricsResult, error) {
	if res, err := t.resultCache.Get(cacheKeyGlobalMetrics); err == nil {
		switch res.(type) {
		case *GlobalMetricsResult:
			return res.(*GlobalMetricsResult), nil
		default:
			return nil, errors.New("couldn't get global metrics from cache, cache contains object of wrong type")
		}
	}

	return nil, errors.New("couldn't get global metrics from cache, cache is empty")
}

func NewGlobalMetrics(conn *sql.DB, cache cache.Cache) *GlobalMetrics {
	return &GlobalMetrics{
		conn:        conn,
		resultCache: cache,
	}
}
