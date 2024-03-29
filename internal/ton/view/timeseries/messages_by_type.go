package timeseries

import (
	"database/sql"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	createTsMessagesByType = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_ts_MessagesByType
	ENGINE = SummingMergeTree() 
	PARTITION BY tuple()
	ORDER BY (Time, WorkchainId, Type, MsgType)
	POPULATE 
	AS
	SELECT
		toStartOfInterval(Time, INTERVAL 5 MINUTE) as Time,
		WorkchainId,
	    Type,
	    Messages.Type as MsgType,
		count() as MessagesCount
	FROM transactions
	ARRAY JOIN Messages
	GROUP BY Time, WorkchainId, Type, MsgType
`
	dropTsMessagesByType = `DROP TABLE _view_ts_MessagesByType`

	selectMessagesByType = `
	SELECT 
       WorkchainId,
       Type,
       MsgType,
	   groupArray(Time),
	   groupArray(MessagesCount)
    FROM (
		SELECT 
			toUInt64(Time) as Time, 
		    WorkchainId,
			Type,
			MsgType,
			sum(MessagesCount) as MessagesCount
		FROM _view_ts_MessagesByType 
		WHERE Time <= now() AND Time >= now()-?
		GROUP BY Time, WorkchainId, Type, MsgType
		ORDER BY Time
	) GROUP BY WorkchainId, Type, MsgType
`
)

type MessagesByTypeResult struct {
	Rows []*MessagesByTypeTimeseries `json:"rows"`
}

type MessagesByTypeTimeseries struct {
	WorkchainId   ton.WorkchainId `json:"workchain_id"`
	Type          string          `json:"type"`
	MsgType       string          `json:"msg_type"`
	Time          []uint64        `json:"time"`
	MessagesCount []uint64        `json:"messages_count"`
}

type MessagesByType struct {
	conn        *sql.DB
	resultCache *cache.WithTimer
}

func (t *MessagesByType) CreateTable() error {
	_, err := t.conn.Exec(createTsMessagesByType)
	return err
}

func (t *MessagesByType) DropTable() error {
	_, err := t.conn.Exec(dropTsMessagesByType)
	return err
}

func (t *MessagesByType) GetMessagesByType() (*MessagesByTypeResult, error) {
	if res, ok := t.resultCache.Get(); ok {
		switch res.(type) {
		case *MessagesByTypeResult:
			return res.(*MessagesByTypeResult), nil
		}
	}

	rows, err := t.conn.Query(selectMessagesByType, []byte("INTERVAL 2 DAY"))
	if err != nil {
		return nil, err
	}

	var resp = &MessagesByTypeResult{
		Rows: make([]*MessagesByTypeTimeseries, 0),
	}

	for rows.Next() {
		row := &MessagesByTypeTimeseries{
			Time:          make([]uint64, 0),
			MessagesCount: make([]uint64, 0),
		}
		if err := rows.Scan(
			&row.WorkchainId,
			&row.Type,
			&row.MsgType,
			&row.Time,
			&row.MessagesCount,
		); err != nil {
			rows.Close()
			return nil, err
		}
		resp.Rows = append(resp.Rows, row)
	}

	rows.Close()

	t.resultCache.Set(resp)

	return resp, nil
}

func NewMessagesByType(conn *sql.DB) *MessagesByType {
	return &MessagesByType{
		conn:        conn,
		resultCache: cache.NewWithTimer(time.Second),
	}
}
