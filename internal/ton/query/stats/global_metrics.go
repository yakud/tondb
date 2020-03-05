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
		sum(TotalTransactions) AS TotalTransactions,
	    sum(TotalBlocks) AS TotalBlocks,
	    sum(TrxLastDay) AS TrxLastDay,
	    sum(TrxLastMonth) AS TrxLastMonth
	FROM (
		SELECT
			count() AS TotalAddr,
			0 AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalTransactions,
		    0 AS TotalBlocks,
		    0 AS TrxLastDay,
		    0 AS TrxLastMonth
		FROM ".inner._view_state_AccountState" FINAL

		UNION ALL

		-- Sum ValueFlow.ToNextBlk by all shards in master head block
		SELECT
			0 AS TotalAddr,
			sum(ValueFlowToNextBlk) AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalTransactions,
		    0 AS TotalBlocks,
		    0 AS TrxLastDay,
		    0 AS TrxLastMonth
		FROM blocks
		PREWHERE (WorkchainId, Shard, SeqNo) IN (
			SELECT
				ShardWorkchainId as WorkchainId,
				Shard,
				ShardSeqNo as SeqNo
			FROM shards_descr
			PREWHERE MasterSeqNo = (SELECT MasterSeqNo FROM shards_descr ORDER BY MasterSeqNo DESC LIMIT 1)

			UNION ALL

			SELECT 
				-1 as WorkchainId,
				9223372036854775808 as Shard,
				(SELECT MasterSeqNo FROM shards_descr ORDER BY MasterSeqNo DESC LIMIT 1) as SeqNo
		)

	 	UNION ALL
		
		SELECT 
		   	0 AS TotalAddr,
		   	0 AS TotalNanogram,
			sum(TotalMessages) AS TotalMessages,
			sum(TotalTransactions) AS TotalTransactions,
			0 AS TotalBlocks,
			0 AS TrxLastDay,
		    0 AS TrxLastMonth
		FROM ".inner._view_feed_TotalTransactionsAndMessages"
		    
		UNION ALL  

		SELECT
		    0 AS TotalAddr,
		    0 AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalTransactions,
		    count() AS TotalBlocks,
    		0 AS TrxLastDay, 
    		0 AS TrxLastMonth
		FROM blocks
		
		UNION ALL  

		SELECT
		    0 AS TotalAddr,
		    0 AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalTransactions,
		    0 AS TotalBlocks,
    		countIf(Time >= (now() - INTERVAL 1 DAY)) AS TrxLastDay, 
    		countIf(Time >= (now() - INTERVAL 1 MONTH)) AS TrxLastMonth
		FROM ".inner._view_feed_TransactionsFeed"
	    WHERE Time >= (now() - INTERVAL 1 MONTH)
	)
`

	cacheKeyGlobalMetrics = "global_metrics"
)

type GlobalMetricsResult struct {
	TotalAddr         uint64 `json:"total_addr"`
	TotalNanogram     uint64 `json:"total_nanogram"`
	TotalBlocks       uint64 `json:"total_blocks"`
	TotalMessages     uint64 `json:"total_messages"`
	TotalTransactions uint64 `json:"total_transactions"`
	TrxLastDay        uint64 `json:"trx_last_day"`
	TrxLastMonth      uint64 `json:"trx_last_month"`
}

type GlobalMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *GlobalMetrics) UpdateQuery() error {
	res := GlobalMetricsResult{}

	row := t.conn.QueryRow(getGlobalMetrics)

	if err := row.Scan(
		&res.TotalAddr,
		&res.TotalNanogram,
		&res.TotalMessages,
		&res.TotalTransactions,
		&res.TotalBlocks,
		&res.TrxLastDay,
		&res.TrxLastMonth,
	); err != nil {
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
