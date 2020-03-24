package filter

import (
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func TypeFromRequest(r *http.Request) (filter.Filter, error) {
	typeQuery, ok := r.URL.Query()["type"]
	if !ok || len(typeQuery) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range typeQuery {
		orFilter.Or(
			filter.NewKV("Type", v),
		)
	}

	return orFilter, nil
}

func TypeFromParams(typeParam *[]string) (filter.Filter, error) {
	if typeParam == nil || len(*typeParam) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range *typeParam {
		orFilter.Or(
			filter.NewKV("Type", v),
		)
	}

	return orFilter, nil
}
