package filter

import (
	"net/http"
	"strings"

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
			filter.NewKV("Hash", strings.TrimSpace(v)),
		)
	}

	return orFilter, nil
}

func HashFromParams(hash *[]string) (filter.Filter, error) {
	if hash == nil || len(*hash) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range *hash {
		orFilter.Or(
			filter.NewKV("Hash", strings.TrimSpace(v)),
		)
	}

	return orFilter, nil
}