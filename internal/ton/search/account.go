package search

import (
	"fmt"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

func (s *Searcher) searchAccount(q string) ([]Result, error) {
	workchainId, addr, err := utils.ParseAccountAddress(q)
	if err != nil {
		return nil, fmt.Errorf("error parse acc query '%s': %w", q, err)
	}

	accFilter := ton.AddrStd{
		WorkchainId: workchainId,
		Addr:        addr,
	}

	if _, err := s.accountStorage.GetAccount(accFilter); err != nil {
		return nil, fmt.Errorf("account not found %s", q)
	}

	return []Result{
		{
			Type: ResultTypeAccount,
			Hint: q,
			Link: "/account?address=" + q,
		},
	}, nil
}
