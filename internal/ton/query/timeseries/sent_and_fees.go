package timeseries

import (
	"database/sql"
	"errors"
	"math"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	getAverageSentAndFees = `
	SELECT
		toStartOfDay(Time) AS Time,
		avg(BlockStatsSentNanograms) AS AvgSent,
		avg(ValueFlowFeesCollected) AS AvgFees
	FROM blocks
	PREWHERE Time >= now() - INTERVAL 30 DAY
	GROUP BY Time
	ORDER BY Time
`

	cacheKeySentAndFees = "sent_and_fees"
)

type SentAndFeesResult struct {
	Time    time.Time `json:"time"`
	AvgSent float64   `json:"avg_sent"`
	AvgFees float64   `json:"avg_fees"`
}

type SentAndFees struct {
	conn        *sql.DB
	resultCache cache.Cache
}

func (t *SentAndFees) UpdateQuery() error {
	res := make([]*SentAndFeesResult, 0, 30)

	rows, err := t.conn.Query(getAverageSentAndFees)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		sentAndFees := &SentAndFeesResult{}
		if err := rows.Scan(&sentAndFees.Time, &sentAndFees.AvgSent, &sentAndFees.AvgFees); err != nil {
			return err
		}

		if math.IsNaN(sentAndFees.AvgFees) {
			sentAndFees.AvgFees = 0
		}

		if math.IsNaN(sentAndFees.AvgSent) {
			sentAndFees.AvgSent = 0
		}

		res = append(res, sentAndFees)
	}

	t.resultCache.Set(cacheKeySentAndFees, &res)

	return nil
}

func (t *SentAndFees) GetSentAndFees() ([]*SentAndFeesResult, error) {
	if res, err := t.resultCache.Get(cacheKeySentAndFees); err == nil {
		switch res.(type) {
		case *[]*SentAndFeesResult:
			return *res.(*[]*SentAndFeesResult), nil
		default:
			return nil, errors.New("couldn't get average sent and fees from cache, cache contains object of wrong type")
		}
	}

	return nil, errors.New("couldn't get average sent and fees from cache, cache is empty")
}

func NewSentAndFees(conn *sql.DB, cache cache.Cache) *SentAndFees {
	return &SentAndFees{
		conn:        conn,
		resultCache: cache,
	}
}
