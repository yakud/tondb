package timeseries

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	createTsMessagesOrdCount = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_ts_MessagesOrdCount
	ENGINE = SummingMergeTree() 
	PARTITION BY tuple()
	ORDER BY (Time, WorkchainId)
	POPULATE 
	AS
	SELECT
		toStartOfInterval(Time, INTERVAL 10 MINUTE) as Time,
		WorkchainId,
	    count() as MessagesCount
	FROM transactions
	ARRAY JOIN Messages
	WHERE 
		Type = 'trans_ord' AND 
		Messages.ValueNanograms > 0
	GROUP BY Time, WorkchainId
`
	dropTsMessagesOrdCount = `DROP TABLE _view_ts_MessagesOrdCount`

	selectMessagesOrdCount = `
	SELECT 
       WorkchainId,
	   groupArray(Time),
	   groupArray(MessagesCountSum)
    FROM (
		SELECT 
			toUInt64(Time) as Time, 
		    WorkchainId,
		    toUInt64(sum(MessagesCount)) as MessagesCountSum
		FROM _view_ts_MessagesOrdCount 
		WHERE Time <= now() AND Time >= now()-?
		GROUP BY Time, WorkchainId
		ORDER BY Time, WorkchainId
	) GROUP BY WorkchainId
`
)

type MessagesOrdCountResult struct {
	Rows []*MessagesOrdCountTimeseries `json:"rows"`
}

type MessagesOrdCountTimeseries struct {
	WorkchainId ton.WorkchainId `json:"workchain_id"`
	Time        []uint64        `json:"time"`
	Count       []uint64        `json:"count"`
}

type MessagesOrdCount struct {
	conn        *sql.DB
	resultCache *cache.WithTimer
}

func (t *MessagesOrdCount) CreateTable() error {
	_, err := t.conn.Exec(createTsMessagesOrdCount)
	return err
}

func (t *MessagesOrdCount) DropTable() error {
	_, err := t.conn.Exec(dropTsMessagesOrdCount)
	return err
}

func (t *MessagesOrdCount) GetMessagesOrdCount() (*tonapi.MessagesOrdCountResult, error) {
	if res, ok := t.resultCache.Get(); ok {
		switch res.(type) {
		case *tonapi.MessagesOrdCountResult:
			return res.(*tonapi.MessagesOrdCountResult), nil
		}
	}

	rows, err := t.conn.Query(selectMessagesOrdCount, []byte("INTERVAL 2 DAY"))
	if err != nil {
		return nil, err
	}

	var resp = &tonapi.MessagesOrdCountResult{
		Rows: make([]tonapi.MessagesOrdCountTimeseries, 0),
	}

	for rows.Next() {
		times := make([]uint64, 0)
		counts := make([]uint64, 0)

		row := tonapi.MessagesOrdCountTimeseries{
			Time:  make([]tonapi.Uint64, 0),
			Count: make([]tonapi.Uint64, 0),
		}

		if err := rows.Scan(
			&row.WorkchainId,
			&times,
			&counts,
		); err != nil {
			rows.Close()
			return nil, err
		}

		for _, v := range times {
			row.Time = append(row.Time, tonapi.Uint64(v))
		}
		for _, v := range counts {
			row.Count = append(row.Count, tonapi.Uint64(v))
		}

		resp.Rows = append(resp.Rows, row)
	}

	rows.Close()

	t.resultCache.Set(resp)

	return resp, nil
}

func NewMessagesOrdCount(conn *sql.DB) *MessagesOrdCount {
	return &MessagesOrdCount{
		conn:        conn,
		resultCache: cache.NewWithTimer(time.Second * 5),
	}
}
