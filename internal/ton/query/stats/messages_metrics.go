package stats

import (
	"database/sql"
	"errors"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

const (
	getTotalTransactionsAndMessages = `
	SELECT
		sum(TotalTransactions) AS TotalTransactions,
		sum(TotalMessages) AS TotalMessages
	FROM ".inner._view_feed_TotalTransactionsAndMessages"
	WHERE %s
`

	getTotalTransactionsAndMessagesForAllChains = `
	SELECT
		sum(TotalTransactions) AS TotalTransactions,
		sum(TotalMessages) AS TotalMessages
	FROM ".inner._view_feed_TotalTransactionsAndMessages"
`

	getTransactionsAndMessagesPerSecond = `
	SELECT 
		avg(MsgCount) AS MPS,
		avg(TrxCount) AS TPS
	FROM (
		SELECT
			sum(TrxCount) as TrxCount,
			sum(MsgCount) as MsgCount
		FROM ".inner._view_feed_MessagesPerSecond"
		WHERE Time > now() - INTERVAL 1 WEEK AND Time <= now() AND %s
		GROUP BY Time
		ORDER BY Time WITH FILL
	)
`

	cacheKeyMessagesMetrics = "messages_metrics"
)

type MessagesMetricsResult struct {
	TotalTransactions  uint64 `json:"total_transactions"`
	TotalMessages      uint64 `json:"total_messages"`

	Tps float64 `json:"tps"`
	Mps float64 `json:"mps"`
}

type MessagesMetrics struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *MessagesMetrics) UpdateQuery() error {
	res := MessagesMetricsResult{}

	row := t.conn.QueryRow(getTotalTransactionsAndMessagesForAllChains)
	if err := row.Scan(&res.TotalTransactions, &res.TotalMessages); err != nil {
		return err
	}

	queryTpsAndMps, _, err := filter.RenderQuery(getTransactionsAndMessagesPerSecond, nil)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryTpsAndMps)
	if err := row.Scan(&res.Tps, &res.Mps); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyMessagesMetrics, &res)

	resWorkchain := MessagesMetricsResult{}
	workchainFilter := filter.NewKV("WorkchainId", 0)

	queryTotalTrxAndMsgs, args, err := filter.RenderQuery(getTotalTransactionsAndMessages, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryTotalTrxAndMsgs, args...)
	if err := row.Scan(&resWorkchain.TotalTransactions, &resWorkchain.TotalMessages); err != nil {
		return err
	}

	queryTpsAndMps, args, err = filter.RenderQuery(getTransactionsAndMessagesPerSecond, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryTpsAndMps, args...)
	if err := row.Scan(&resWorkchain.Tps, &resWorkchain.Mps); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyMessagesMetrics + "0", &resWorkchain)

	resMasterchain := MessagesMetricsResult{}
	workchainFilter = filter.NewKV("WorkchainId", -1)

	queryTotalTrxAndMsgs, args, err = filter.RenderQuery(getTotalTransactionsAndMessages, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryTotalTrxAndMsgs, args...)
	if err := row.Scan(&resMasterchain.TotalTransactions, &resMasterchain.TotalMessages); err != nil {
		return err
	}

	queryTpsAndMps, args, err = filter.RenderQuery(getTransactionsAndMessagesPerSecond, workchainFilter)
	if err != nil {
		return err
	}
	row = t.conn.QueryRow(queryTpsAndMps, args...)
	if err := row.Scan(&resMasterchain.Tps, &resMasterchain.Mps); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyMessagesMetrics + "-1", &resMasterchain)

	return nil
}

func (t *MessagesMetrics) GetMessagesMetrics(workchainId string) (*MessagesMetricsResult, error) {
	if res, err := t.resultCache.Get(cacheKeyMessagesMetrics + workchainId); err == nil {
		switch res.(type) {
		case *MessagesMetricsResult:
			return res.(*MessagesMetricsResult), nil
		default:
			return nil, errors.New("couldn't get messages metrics from cache, cache contains object of wrong type")
		}
	}

	return nil, errors.New("couldn't get messages metrics from cache, wrong workchainId specified or cache is empty")
}

func NewMessagesMetrics(conn *sql.DB, cache cache.Cache) *MessagesMetrics {
	return &MessagesMetrics{
		conn:        conn,
		resultCache: cache,
	}
}

