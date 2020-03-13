package stats

import (
	"database/sql"
	"errors"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/internal/utils"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	selectTopWhales = `
	SELECT
	  	concat(toString(WorkchainId),':',Addr) as Addr,
		BalanceNanogram
	FROM ".inner._view_state_AccountState" FINAL
	%s
	ORDER BY BalanceNanogram DESC
	LIMIT %d %s
`
	whereWorkchainId = "WHERE WorkchainId = %d"

	queryOffset = "OFFSET %d"

	cacheKeyTopWhales = "top_whales"

	WhalesDefaultCacheLimit = 10000

	WhalesDefaultPageLimit = 50
)

type Whale struct {
	AddrRaw     string `json:"addr"`
	AddrUf      string `json:"addr_uf"`

	BalanceNanogram          uint64 `json:"balance_nanogram"`
	BalancePercentageOfTotal float64 `json:"balance_percentage_of_total"`
}

type TopWhales []Whale

type GetTopWhales struct {
	conn          *sql.DB
	resultCache   cache.Cache
	globalMetrics *GlobalMetrics
}

func (q *GetTopWhales) UpdateQuery() error {
	// Not using filter here because it is not giving such flexibility as just fmt.Sprintf() in this case
	res, err := q.queryTopWhales(fmt.Sprintf(selectTopWhales, "", WhalesDefaultCacheLimit, ""))
	if err != nil {
		return err
	}

	q.resultCache.Set(cacheKeyTopWhales, res)

	resWorkchain, err := q.queryTopWhales(fmt.Sprintf(selectTopWhales, fmt.Sprintf(whereWorkchainId, 0), WhalesDefaultCacheLimit, ""))
	if err != nil {
		return err
	}

	q.resultCache.Set(cacheKeyTopWhales + "0", resWorkchain)

	resMasterchain, err := q.queryTopWhales(fmt.Sprintf(selectTopWhales, fmt.Sprintf(whereWorkchainId, -1), WhalesDefaultCacheLimit, ""))
	if err != nil {
		return err
	}

	q.resultCache.Set(cacheKeyTopWhales + "-1", resMasterchain)

	return nil
}

func (q *GetTopWhales) GetTopWhales(workchainId int32, limit uint32, offset uint32) (*TopWhales, error) {
	if limit <= 0 {
		limit = WhalesDefaultPageLimit
	}

	if limit + offset > WhalesDefaultCacheLimit {
		workchainFilter := ""
		if workchainId != feed.EmptyWorkchainId {
			workchainFilter = fmt.Sprintf(whereWorkchainId, workchainId)
		}

		return q.queryTopWhales(fmt.Sprintf(selectTopWhales, workchainFilter, limit, fmt.Sprintf(queryOffset, offset)))
	}

	workchainIdStr := ""
	if workchainId != feed.EmptyWorkchainId {
		workchainIdStr = strconv.Itoa(int(workchainId))
	}
	if res, err := q.resultCache.Get(cacheKeyTopWhales + workchainIdStr); err == nil {
		switch res.(type) {
		case *TopWhales:
			resPaginated := make(TopWhales, 0, limit)
			resPaginated = append(resPaginated, (*res.(*TopWhales))[offset:offset+limit]...)
			return &resPaginated, nil
		default:
			return nil, errors.New("couldn't get top whales from cache, cache contains object of wrong type")
		}
	}

	return nil, errors.New("couldn't get top whales from cache, wrong workchainId specified or cache is empty")
}

func (q *GetTopWhales) queryTopWhales(query string) (*TopWhales, error) {
	globalMetrics, err := q.globalMetrics.GetGlobalMetrics()
	if err != nil {
		return nil, err
	}
	totalNanogram := float64(globalMetrics.TotalNanogram)

	rows, err := q.conn.Query(query)
	if err != nil {
		return nil, err
	}

	var resp = make(TopWhales, 0, WhalesDefaultCacheLimit)
	for rows.Next() {
		whale := Whale{}

		if err := rows.Scan(&whale.AddrRaw, &whale.BalanceNanogram); err != nil {
			return nil, err
		}
		if whale.AddrUf, err = utils.ConvertRawToUserFriendly(whale.AddrRaw, utils.UserFriendlyAddrDefaultTag); err != nil {
			// Maybe we shouldn't fail here?
			return nil, err
		}

		whale.BalancePercentageOfTotal = float64(whale.BalanceNanogram) / totalNanogram

		resp = append(resp, whale)
	}

	return &resp, nil
}

func NewGetTopWhales(conn *sql.DB, cache cache.Cache, metrics *GlobalMetrics) *GetTopWhales {
	return &GetTopWhales{
		conn:          conn,
		resultCache:   cache,
		globalMetrics: metrics,
	}
}
