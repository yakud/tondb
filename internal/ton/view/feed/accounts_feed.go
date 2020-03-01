package feed

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	DefaultAccountsLimit = 50
	MaxAccountsLimit     = 500

	querySelectAccounts = `
	SELECT
	    WorkchainId,
	    Addr,
	    BalanceNanogram
	FROM ".inner._view_state_AccountState" FINAL
	PREWHERE 
	     if(? == bitShiftLeft(toInt32(-1), 31), 1, WorkchainId = ?)
	LIMIT ?, ?
`
)

type AccountInFeed struct {
	WorkchainId int32  `json:"workchain_id"`
	Addr        string `json:"addr"`
	AddrUf      string `json:"addr_uf"`

	BalanceNanogram 		 uint64  `json:"balance_nanogram"`
	BalancePercentageOfTotal float64 `json:"balance_percentage_of_total"`
}

type AccountsFeed struct {
	conn          *sql.DB
	globalMetrics *stats.GlobalMetrics
}

func (t *AccountsFeed) SelectAccounts(wcId int32, limit uint16, offset uint32) ([]*AccountInFeed, error) {
	if limit == 0 {
		limit = DefaultAccountsLimit
	}
	if limit > MaxAccountsLimit {
		limit = MaxAccountsLimit
	}

	rows, err := t.conn.Query(querySelectAccounts, wcId, wcId, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feed []*AccountInFeed

	for rows.Next() {
		acc := &AccountInFeed{}
		if err = rows.Scan(&acc.WorkchainId, &acc.Addr, &acc.BalanceNanogram); err != nil {
			return nil, err
		}

		globalMetrics, err := t.globalMetrics.GetGlobalMetrics()
		if err != nil {
			return nil, err
		}

		acc.BalancePercentageOfTotal = float64(acc.BalanceNanogram) / float64(globalMetrics.TotalNanogram)

		if acc.AddrUf, err = utils.ComposeRawAndConvertToUserFriendly(acc.WorkchainId, acc.Addr); err != nil {
			return nil, err
		}

		feed = append(feed, acc)
	}

	return feed, nil
}

func NewAccountsFeed(conn *sql.DB, globalMetrics *stats.GlobalMetrics) *AccountsFeed {
	return &AccountsFeed{
		conn:          conn,
		globalMetrics: globalMetrics,
	}
}
