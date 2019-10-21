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
			filter.NewKV("MessageSrcAddr", v),
			filter.NewKV("MessageDestAddr", v),
		)
	}

	return orFilter, nil
}
