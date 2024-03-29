package stats

import (
	"database/sql"
	"errors"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

const (
	getTotalAddrAndNanogram = `
	SELECT 
		sum(TotalAddr),
		sum(TotalNanogram)
	FROM (
		SELECT 
			count() AS TotalAddr, 
			0 AS TotalNanogram
		FROM ".inner._view_state_AccountState"
		FINAL
		WHERE %s
	
		UNION ALL
	
		SELECT
			0  AS TotalAddr,
			sum(ValueFlowToNextBlk) AS TotalNanogram
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
		) AND %s
	)
`

	getDailyActiveAccounts = `
	SELECT
		count() as ActiveAddr
	FROM ".inner._view_state_AccountState"
	FINAL
	WHERE Time > now() - INTERVAL 1 DAY AND %s;
`

	getMonthlyActiveAccounts = `
	SELECT
		count() as ActiveAddr
	FROM ".inner._view_state_AccountState"
	FINAL
	WHERE Time > now() - INTERVAL 1 MONTH AND %s;
`

	cacheKeyAddressesMetrics = "addresses_metrics"
)

type AddressesMetricsResult struct {
	TotalAddr     uint64 `json:"total_addr"`
	TotalNanogram uint64 `json:"total_nanogram"`
	DailyActive   uint64 `json:"daily_active"`
	MonthlyActive uint64 `json:"monthly_active"`
}

type AddressesMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *AddressesMetrics) UpdateQuery() error {
	res := AddressesMetricsResult{}

	filterAll := filter.NewKV("1", 1)
	queryGetTotalAddrAndNanogram, args, err := filter.RenderQuery(getTotalAddrAndNanogram, filterAll, filterAll)
	if err != nil {
		return err
	}
	row := t.conn.QueryRow(queryGetTotalAddrAndNanogram, args...)
	if err := row.Scan(&res.TotalAddr, &res.TotalNanogram); err != nil {
		return err
	}

	queryGetDailyActiveAccounts, _, err := filter.RenderQuery(getDailyActiveAccounts, nil)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetDailyActiveAccounts)
	if err := row.Scan(&res.DailyActive); err != nil {
		return err
	}

	queryGetMonthlyActiveAccounts, _, err := filter.RenderQuery(getMonthlyActiveAccounts, nil)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetMonthlyActiveAccounts)
	if err := row.Scan(&res.MonthlyActive); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyAddressesMetrics, &res)

	resWorkchain := AddressesMetricsResult{}
	workchainFilter := filter.NewKV("WorkchainId", 0)

	queryGetTotalAddrAndNanogram, args, err = filter.RenderQuery(getTotalAddrAndNanogram, workchainFilter, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetTotalAddrAndNanogram, args...)
	if err := row.Scan(&resWorkchain.TotalAddr, &resWorkchain.TotalNanogram); err != nil {
		return err
	}

	queryGetDailyActiveAccounts, args, err = filter.RenderQuery(getDailyActiveAccounts, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetDailyActiveAccounts, args...)
	if err := row.Scan(&resWorkchain.DailyActive); err != nil {
		return err
	}

	queryGetMonthlyActiveAccounts, args, err = filter.RenderQuery(getMonthlyActiveAccounts, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetMonthlyActiveAccounts, args...)
	if err := row.Scan(&resWorkchain.MonthlyActive); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyAddressesMetrics+"0", &resWorkchain)

	resMasterchain := AddressesMetricsResult{}
	workchainFilter = filter.NewKV("WorkchainId", -1)

	queryGetTotalAddrAndNanogram, args, err = filter.RenderQuery(getTotalAddrAndNanogram, workchainFilter, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetTotalAddrAndNanogram, args...)
	if err := row.Scan(&resMasterchain.TotalAddr, &resMasterchain.TotalNanogram); err != nil {
		return err
	}

	queryGetDailyActiveAccounts, args, err = filter.RenderQuery(getDailyActiveAccounts, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetDailyActiveAccounts, args...)
	if err := row.Scan(&resMasterchain.DailyActive); err != nil {
		return err
	}

	queryGetMonthlyActiveAccounts, args, err = filter.RenderQuery(getMonthlyActiveAccounts, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryGetMonthlyActiveAccounts, args...)
	if err := row.Scan(&resMasterchain.MonthlyActive); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyAddressesMetrics+"-1", &resMasterchain)

	return nil
}

func (t *AddressesMetrics) GetAddressesMetrics(workchainId string) (*AddressesMetricsResult, error) {
	if res, err := t.resultCache.Get(cacheKeyAddressesMetrics + workchainId); err == nil {
		switch res.(type) {
		case *AddressesMetricsResult:
			return res.(*AddressesMetricsResult), nil
		default:
			return nil, errors.New("couldn't get addresses metrics from cache, cache contains object of wrong type")
		}
	}

	return nil, errors.New("couldn't get addresses metrics from cache, wrong workchainId specified or cache is empty")
}

func NewAddressesMetrics(conn *sql.DB, cache cache.Cache) *AddressesMetrics {
	return &AddressesMetrics{
		conn:        conn,
		resultCache: cache,
	}
}
