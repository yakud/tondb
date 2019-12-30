package filter

import (
	"errors"
	"net/http"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func AccountFilterFromRequest(r *http.Request, field string) (*filter.Account, error) {
	addr, ok := r.URL.Query()[field]
	if !ok {
		return nil, errors.New("address is empty")
	}
	if len(addr) != 1 {
		return nil, errors.New("wrong count address field should be exactly one")
	}

	accAddr, err := ton.ParseAccountAddress(strings.TrimSpace(addr[0]))
	if err != nil {
		return nil, err
	}

	return filter.NewAccount(accAddr), nil
}
