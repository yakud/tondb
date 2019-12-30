package filter

import (
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func MessageAddrFromRequest(r *http.Request) (filter.Filter, error) {
	addrQuery, ok := r.URL.Query()["addr"]
	if !ok || len(addrQuery) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range addrQuery {
		orFilter.Or(
			filter.NewArrayHas("Messages.SrcAddr", v),
			filter.NewArrayHas("Messages.DestAddr", v),
		)
	}

	return orFilter, nil
}
