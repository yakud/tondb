package stats

import (
	"database/sql"
	"errors"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/internal/utils"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
)

const (
	selectTopWhales = `
	SELECT
	  	concat(toString(WorkchainId),':',Addr) as Addr,
		BalanceNanogram
	FROM ".inner._view_state_AccountState" FINAL
	WHERE %s
	ORDER BY BalanceNanogram DESC
	LIMIT ?
`

	cacheKeyTopWhales = "top_whales"

	WhalesDefaultCacheLimit = 10000

	WhalesDefaultPageLimit = 50
)

type Whale struct {
	AddrRaw string `json:"addr"`
	AddrUf  string `json:"addr_uf"`

	BalanceNanogram          uint64  `json:"balance_nanogram"`
	BalancePercentageOfTotal float64 `json:"balance_percentage_of_total"`
}

type TopWhales []Whale

type GetTopWhales struct {
	conn            *sql.DB
	resultCache     cache.Cache
	addressesMetric *AddressesMetrics
}

func (q *GetTopWhales) UpdateQuery() error {
	// Not using filter here because it is not giving such flexibility as just fmt.Sprintf() in this case
	res, err := q.queryTopWhales(-2)
	if err != nil {
		return err
	}

	q.resultCache.Set(cacheKeyTopWhales, res)

	resWorkchain, err := q.queryTopWhales(0)
	if err != nil {
		return err
	}

	q.resultCache.Set(cacheKeyTopWhales+"0", resWorkchain)

	resMasterchain, err := q.queryTopWhales(-1)
	if err != nil {
		return err
	}

	q.resultCache.Set(cacheKeyTopWhales+"-1", resMasterchain)

	return nil
}

func (q *GetTopWhales) GetTopWhales(workchainId int32, limit uint32, offset uint32) (*TopWhales, error) {
	if limit <= 0 {
		limit = WhalesDefaultPageLimit
	}

	if limit+offset > WhalesDefaultCacheLimit {
		// empty result after over limit
		return &TopWhales{}, nil
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

func (q *GetTopWhales) queryTopWhales(workchainId int32) (*TopWhales, error) {
	var err error
	var addrMetrics *AddressesMetricsResult

	var f filter.Filter
	if workchainId == -2 {
		f = filter.NewKV("1", 1)

		if addrMetrics, err = q.addressesMetric.GetAddressesMetrics(""); err != nil {
			return nil, err
		}
	} else {
		f = filter.NewKV("WorkchainId", workchainId)

		if addrMetrics, err = q.addressesMetric.GetAddressesMetrics(strconv.Itoa(int(workchainId))); err != nil {
			return nil, err
		}
	}

	query, args, err := filter.RenderQuery(selectTopWhales, f)
	if err != nil {
		return nil, err
	}

	args = append(args, WhalesDefaultCacheLimit)

	rows, err := q.conn.Query(query, args...)
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

		whale.BalancePercentageOfTotal = float64(whale.BalanceNanogram) / float64(addrMetrics.TotalNanogram)

		resp = append(resp, whale)
	}

	return &resp, nil
}

func NewGetTopWhales(conn *sql.DB, cache cache.Cache, addressesMetric *AddressesMetrics) *GetTopWhales {
	return &GetTopWhales{
		conn:            conn,
		resultCache:     cache,
		addressesMetric: addressesMetric,
	}
}
