package stats

import (
	"database/sql"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	selectTopWhales = `
	SELECT
	  	concat(toString(WorkchainId),':',Addr) as Addr,
		toDecimal128(BalanceNanogram, 10) * toDecimal128(0.000000001, 10) as BalanceGram
	FROM ".inner._view_state_AccountState" FINAL
	ORDER BY BalanceNanogram DESC
	LIMIT 50
`
)

type Whale struct {
	Addr        string `json:"addr"`
	BalanceGram string `json:"balance_gram"`
}

type TopWhales []Whale

type GetTopWhales struct {
	conn        *sql.DB
	resultCache *cache.WithTimer
}

func (q *GetTopWhales) GetTopWhales() (*TopWhales, error) {
	if res, ok := q.resultCache.Get(); ok {
		switch res.(type) {
		case *TopWhales:
			return res.(*TopWhales), nil
		}
	}

	rows, err := q.conn.Query(selectTopWhales)
	if err != nil {
		return nil, err
	}

	var resp = make(TopWhales, 0, 50)
	for rows.Next() {
		whale := Whale{}

		if err := rows.Scan(&whale.Addr, &whale.BalanceGram); err != nil {
			return nil, err
		}

		resp = append(resp, whale)
	}

	q.resultCache.Set(&resp)

	return &resp, nil
}

func NewGetTopWhales(conn *sql.DB) *GetTopWhales {
	return &GetTopWhales{
		conn:        conn,
		resultCache: cache.NewWithTimer(time.Second * 10),
	}
}
