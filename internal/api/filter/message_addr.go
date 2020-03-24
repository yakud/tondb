package filter

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"net/http"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func MessageAddrFromRequest(r *http.Request) (filter.Filter, error) {
	addrQuery, ok := r.URL.Query()["addr"]
	if !ok || len(addrQuery) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range addrQuery {
		addr, err := ton.ParseAccountAddress(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		orFilter.Or(
			filter.NewArrayHas("Messages.SrcAddr", addr),
			filter.NewArrayHas("Messages.DestAddr", addr),
		)
	}

	return orFilter, nil
}

func MessageAddrFromParam(addr *[]string) (filter.Filter, error) {
	if addr == nil || len(*addr) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range *addr {
		addr, err := ton.ParseAccountAddress(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		orFilter.Or(
			filter.NewArrayHas("Messages.SrcAddr", addr),
			filter.NewArrayHas("Messages.DestAddr", addr),
		)
	}

	return orFilter, nil
}
