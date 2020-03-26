package stats

import (
	"database/sql"
	"errors"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	getGlobalMetrics = `
	SELECT
		sum(TotalAddr) AS TotalAddr,
	    sum(TotalNanogram) AS TotalNanogram,
	    sum(TotalMessages) AS TotalMessages,
		sum(TotalTransactions) AS TotalTransactions,
	    sum(TotalBlocks) AS TotalBlocks
	FROM (
		SELECT
			count() AS TotalAddr,
			0 AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalTransactions,
		    0 AS TotalBlocks
		FROM ".inner._view_state_AccountState" FINAL

		UNION ALL

		-- Sum ValueFlow.ToNextBlk by all shards in master head block
		SELECT
			0 AS TotalAddr,
			sum(ValueFlowToNextBlk) AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalTransactions,
		    0 AS TotalBlocks
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
			0 AS TotalBlocks
		FROM ".inner._view_feed_TotalTransactionsAndMessages"
		    
		UNION ALL  

		SELECT
		    0 AS TotalAddr,
		    0 AS TotalNanogram,
		    0 AS TotalMessages,
		    0 AS TotalTransactions,
		    count() AS TotalBlocks
		FROM blocks
	)
`

	cacheKeyGlobalMetrics = "global_metrics"
)

type GlobalMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *GlobalMetrics) UpdateQuery() error {
	res := tonapi.GlobalMetrics{}

	row := t.conn.QueryRow(getGlobalMetrics)

	if err := row.Scan(
		&res.TotalAddr,
		&res.TotalNanogram,
		&res.TotalMessages,
		&res.TotalTransactions,
		&res.TotalBlocks,
	); err != nil {
		return err
	}

	// Not handling the error because it's quite useless in this implementation of Cache interface (cache.Background).
	// Set function returns error in Cache interface because in some implementations of it errors may be somehow valuable.
	t.resultCache.Set(cacheKeyGlobalMetrics, &res)

	return nil
}

func (t *GlobalMetrics) GetGlobalMetrics() (*tonapi.GlobalMetrics, error) {
	if res, err := t.resultCache.Get(cacheKeyGlobalMetrics); err == nil {
		switch res.(type) {
		case *tonapi.GlobalMetrics:
			return res.(*tonapi.GlobalMetrics), nil
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
