package stats

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
	"log"
	"time"
)

const (
	getTotalAddrAndGram = `
	SELECT
		count() AS TotalAddr,
		toDecimal128(sum(BalanceNanogram), 10) * toDecimal128(0.000000001, 10) AS TotalGram
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
	TotalGrm      string `json:"total_grm"`
	TotalBlocks   uint64 `json:"total_blocks"`
	TotalMessages uint64 `json:"total_messages"`
}

type GlobalMetrics struct {
	conn        *sql.DB
	resultCache *cache.Background
}

func Updater(conn *sql.DB) interface{} {
	res := GlobalMetricsResult{}

	row := conn.QueryRow(getTotalAddrAndGram)
	if err := row.Scan(&res.TotalAddr, &res.TotalGrm); err != nil {
		log.Println("Got err while updating global metrics cache querying total addresses and GRM. err: ", err)
		return nil
	}

	row = conn.QueryRow(getTotalBlocks)
	if err := row.Scan(&res.TotalBlocks); err != nil {
		log.Println("Got err while updating global metrics cache querying total blocks. err: ", err)
		return nil
	}

	row = conn.QueryRow(getTotalMessages)
	if err := row.Scan(&res.TotalMessages); err != nil {
		log.Println("Got err while updating global metrics cache querying total messages. err: ", err)
		return nil
	}

	return &res
}

func (t *GlobalMetrics) GetGlobalMetrics() (*GlobalMetricsResult, error) {
	if res, ok := t.resultCache.Get(); ok {
		switch res.(type) {
		case *GlobalMetricsResult:
			return res.(*GlobalMetricsResult), nil
		default:
			return nil, fmt.Errorf("couldn't get global metrics from cache, cache contains object of wrong type")
		}
	}

	return nil, fmt.Errorf("couldn't get global metrics from cache, cache is empty")
}

func NewGlobalMetrics(conn *sql.DB, dur time.Duration, ctx context.Context) *GlobalMetrics {
	return &GlobalMetrics{
		conn:        conn,
		resultCache: cache.NewBackground(Updater, conn, dur, ctx),
	}
}
