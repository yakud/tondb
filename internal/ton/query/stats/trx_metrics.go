package stats

import (
	"database/sql"
	"errors"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

const (
	getTrxMetrics = `
	SELECT
    		countIf(Time >= (now() - INTERVAL 1 DAY)) AS TrxLastDay, 
    		countIf(Time >= (now() - INTERVAL 1 MONTH)) AS TrxLastMonth
		FROM ".inner._view_feed_TransactionsFeed"
	    WHERE Time >= (now() - INTERVAL 1 MONTH) AND %s
`
	cacheKeyTrxMetrics = "trx_metrics"
)

type TrxMetricsResult struct {
	TrxLastDay    uint64 `json:"trx_last_day"`
	TrxLastMonth  uint64 `json:"trx_last_month"`
}

type TrxMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *TrxMetrics) UpdateQuery() error {
	res := TrxMetricsResult{}

	queryGetTrxMetrics, _, err := filter.RenderQuery(getTrxMetrics, nil)
	if err != nil {
		return err
	}
	row := t.conn.QueryRow(queryGetTrxMetrics)
	if err := row.Scan(&res.TrxLastDay, &res.TrxLastMonth); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyTrxMetrics, &res)

	resWorkchain := TrxMetricsResult{}
	workchainFilter := filter.NewKV("WorkchainId", 0)

	queryGetTrxMetrics, args, err := filter.RenderQuery(getTrxMetrics, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetTrxMetrics, args...)
	if err := row.Scan(&res.TrxLastDay, &res.TrxLastMonth); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyTrxMetrics+"0", &resWorkchain)

	resMasterchain := TrxMetricsResult{}
	workchainFilter = filter.NewKV("WorkchainId", -1)

	queryGetTrxMetrics, args, err = filter.RenderQuery(getTrxMetrics, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetTrxMetrics, args...)
	if err := row.Scan(&res.TrxLastDay, &res.TrxLastMonth); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyTrxMetrics+"-1", &resMasterchain)

	return nil
}

func (t *TrxMetrics) GetTrxMetrics(workchainId string) (*TrxMetricsResult, error) {
	if res, err := t.resultCache.Get(cacheKeyTrxMetrics + workchainId); err == nil {
		switch res.(type) {
		case *TrxMetricsResult:
			return res.(*TrxMetricsResult), nil
		default:
			return nil, errors.New("couldn't get trx metrics from cache, cache contains object of wrong type")
		}
	}

	return nil, errors.New("couldn't get trx metrics from cache, wrong workchainId specified or cache is empty")
}

func NewTrxMetrics(conn *sql.DB, cache cache.Cache) *TrxMetrics {
	return &TrxMetrics{
		conn:        conn,
		resultCache: cache,
	}
}
