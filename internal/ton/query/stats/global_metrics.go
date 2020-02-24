package stats

import (
	"database/sql"
	"errors"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	cacheKey = "global_metrics"

	getTotalAddrAndGram = `
	SELECT
		count() AS TotalAddr,
		sum(BalanceNanogram) AS TotalNanogram
    FROM ".inner._view_state_AccountState"
	FINAL
`

	getTotalBlocks = `
	SELECT count() FROM blocks;
`

	getTotalMessages = `
	SELECT count() FROM ".inner._view_feed_MessagesFeedGlobal"
`
)

type GlobalMetricsResult struct {
	TotalAddr     uint64 `json:"total_addr"`
	TotalNanogram uint64 `json:"total_nanogram"`
	TotalBlocks   uint64 `json:"total_blocks"`
	TotalMessages uint64 `json:"total_messages"`
}

type GlobalMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *GlobalMetrics) UpdateQuery() error {
	res := GlobalMetricsResult{}

	row := t.conn.QueryRow(getTotalAddrAndGram)
	if err := row.Scan(&res.TotalAddr, &res.TotalNanogram); err != nil {
		return err
	}

	row = t.conn.QueryRow(getTotalBlocks)
	if err := row.Scan(&res.TotalBlocks); err != nil {
		return err
	}

	row = t.conn.QueryRow(getTotalMessages)
	if err := row.Scan(&res.TotalMessages); err != nil {
		return err
	}

	// Not handling the error because it's quite useless in this implementation of Cache interface (cache.Background).
	// Set function returns error in Cache interface because in some implementations of it errors may be somehow valuable.
	t.resultCache.Set(cacheKey, &res)

	return nil
}

func (t *GlobalMetrics) GetGlobalMetrics() (*GlobalMetricsResult, error) {
	if res, err := t.resultCache.Get(cacheKey); err == nil {
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
