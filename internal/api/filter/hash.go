package filter

import (
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func HashFromRequest(r *http.Request) (filter.Filter, error) {
	typeQuery, ok := r.URL.Query()["hash"]
	if !ok || len(typeQuery) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range typeQuery {
		orFilter.Or(
			filter.NewKV("Hash", v),
		)
	}

	return orFilter, nil
}
