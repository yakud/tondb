package stats

import (
	"database/sql"
	"errors"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	createMessagesPerSecondView = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_MessagesPerSecond
	ENGINE = MergeTree()
	PARTITION BY toYYYYMM(Time)
	ORDER BY (Time, WorkchainId)
	SETTINGS index_granularity = 128, index_granularity_bytes = 0
	POPULATE AS
	SELECT 
    	WorkchainId, 
    	Shard, 
    	SeqNo, 
    	Time, 
    	count() AS MsgCount
	FROM transactions
	ARRAY JOIN Messages
	GROUP BY 
    	WorkchainId, 
    	Shard, 
    	SeqNo, 
    	Time
`

	dropMessagesPerSecondView = `DROP TABLE _view_feed_MessagesPerSecond`

	getTotalTransactions = `SELECT sum(Count) FROM ".inner._view_feed_TransactionFeesFeed" %s` // It turned out to be faster then SELECT count() FROM transactions

	getTotalMessages = `SELECT sum(MsgCount) FROM ".inner._view_feed_MessagesPerSecond" %s`

	getTransactionsPerSecond = `
	SELECT
		avg(TrxCount) AS TPS
	FROM (
		SELECT 
			sum(Count) as TrxCount
		FROM ".inner._view_feed_TransactionFeesFeed"
		%s %s
		GROUP BY WorkchainId, Time
		ORDER BY Time WITH FILL
	)
`

	getMessagesPerSecond = `
	SELECT 
		avg(MsgCount) AS MPS
	FROM (
		SELECT
			sum(MsgCount) as MsgCount
		FROM ".inner._view_feed_MessagesPerSecond"
		%s %s
		GROUP BY WorkchainId, Time
		ORDER BY Time WITH FILL
	)
`

	timeBetweenInterval = "WHERE Time > now() - INTERVAL %d %s AND Time <= now()"

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

func (t *MessagesMetrics) CreateTable() error {
	_, err := t.conn.Exec(createMessagesPerSecondView)
	return err
}

func (t *MessagesMetrics) DropTable() error {
	_, err := t.conn.Exec(dropMessagesPerSecondView)
	return err
}

func (t *MessagesMetrics) UpdateQuery() error {
	res := MessagesMetricsResult{}

	row := t.conn.QueryRow(fmt.Sprintf(getTotalTransactions, ""))
	if err := row.Scan(&res.TotalTransactions); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getTotalMessages, ""))
	if err := row.Scan(&res.TotalMessages); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getTransactionsPerSecond, fmt.Sprintf(timeBetweenInterval, 1, intervalYear), ""))
	if err := row.Scan(&res.Tps); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getMessagesPerSecond, fmt.Sprintf(timeBetweenInterval, 1, intervalYear), ""))
	if err := row.Scan(&res.Mps); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyMessagesMetrics, &res)

	resWorkchain := MessagesMetricsResult{}

	row = t.conn.QueryRow(fmt.Sprintf(getTotalTransactions, fmt.Sprintf(workchainIdPrewhere, 0)))
	if err := row.Scan(&resWorkchain.TotalTransactions); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getTotalMessages, fmt.Sprintf(workchainIdPrewhere, 0)))
	if err := row.Scan(&resWorkchain.TotalMessages); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getTransactionsPerSecond, fmt.Sprintf(timeBetweenInterval, 1, intervalYear), fmt.Sprintf(workchainIdAnd, 0)))
	if err := row.Scan(&resWorkchain.Tps); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getMessagesPerSecond, fmt.Sprintf(timeBetweenInterval, 1, intervalYear), fmt.Sprintf(workchainIdAnd, 0)))
	if err := row.Scan(&resWorkchain.Mps); err != nil {
		return err
	}

	t.resultCache.Set(cacheKeyMessagesMetrics + "0", &resWorkchain)

	resMasterchain := MessagesMetricsResult{}

	row = t.conn.QueryRow(fmt.Sprintf(getTotalTransactions, fmt.Sprintf(workchainIdPrewhere, -1)))
	if err := row.Scan(&resMasterchain.TotalTransactions); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getTotalMessages, fmt.Sprintf(workchainIdPrewhere, -1)))
	if err := row.Scan(&resMasterchain.TotalMessages); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getTransactionsPerSecond, fmt.Sprintf(timeBetweenInterval, 1, intervalYear), fmt.Sprintf(workchainIdAnd, -1)))
	if err := row.Scan(&resMasterchain.Tps); err != nil {
		return err
	}

	row = t.conn.QueryRow(fmt.Sprintf(getMessagesPerSecond, fmt.Sprintf(timeBetweenInterval, 1, intervalYear), fmt.Sprintf(workchainIdAnd, -1)))
	if err := row.Scan(&resMasterchain.Mps); err != nil {
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

