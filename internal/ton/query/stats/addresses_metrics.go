package stats

import (
	"database/sql"
	"errors"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	getTotalAddrAndNanogram = `
	SELECT 
    	count() AS TotalAddr, 
    	sum(BalanceNanogram) AS TotalNanogram
	FROM ".inner._view_state_AccountState"
	FINAL %s
`

	getActiveAccounts = `
	SELECT
		count() as ActiveAddr
	FROM ".inner._view_state_AccountState"
	FINAL
	WHERE Time > now() - INTERVAL %d %s %s;
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

	row := t.conn.QueryRow(fmt.Sprintf(getTotalAddrAndNanogram, ""))
	if err := row.Scan(&res.TotalAddr, &res.TotalNanogram); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getActiveAccounts, 1, intervalDay, ""))
	if err := row.Scan(&res.DailyActive); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getActiveAccounts, 1, intervalMonth, ""))
	if err := row.Scan(&res.MonthlyActive); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyAddressesMetrics, &res)

	resWorkchain := AddressesMetricsResult{}

	row = t.conn.QueryRow(fmt.Sprintf(getTotalAddrAndNanogram, fmt.Sprintf(workchainIdPrewhere, 0)))
	if err := row.Scan(&resWorkchain.TotalAddr, &resWorkchain.TotalNanogram); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getActiveAccounts,  1, intervalDay, fmt.Sprintf(workchainIdAnd, 0)))
	if err := row.Scan(&resWorkchain.DailyActive); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getActiveAccounts,  1, intervalMonth, fmt.Sprintf(workchainIdAnd, 0)))
	if err := row.Scan(&resWorkchain.MonthlyActive); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyAddressesMetrics + "0", &resWorkchain)

	resMasterchain := AddressesMetricsResult{}

	row = t.conn.QueryRow(fmt.Sprintf(getTotalAddrAndNanogram, fmt.Sprintf(workchainIdPrewhere, -1)))
	if err := row.Scan(&resMasterchain.TotalAddr, &resMasterchain.TotalNanogram); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getActiveAccounts,  1, intervalDay, fmt.Sprintf(workchainIdAnd, -1)))
	if err := row.Scan(&resMasterchain.DailyActive); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getActiveAccounts,  1, intervalMonth, fmt.Sprintf(workchainIdAnd, -1)))
	if err := row.Scan(&resMasterchain.MonthlyActive); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyAddressesMetrics + "-1", &resMasterchain)

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

