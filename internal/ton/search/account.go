package search

import (
	"fmt"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

func (s *Searcher) searchAccount(q string) ([]Result, error) {
	workchainId, addr, err := utils.ParseAccountAddress(q)
	if err != nil {
		return nil, err
	}

	accFilter := filter.NewAccount(ton.AddrStd{
		WorkchainId: workchainId,
		Addr:        addr,
	})

	if _, err := s.accountStorage.GetAccount(accFilter); err != nil {
		return nil, fmt.Errorf("account not found")
	}

	return []Result{
		{
			Type: ResultTypeAccount,
			Hint: q,
			Link: "/account?address=" + q,
		},
	}, nil
}
