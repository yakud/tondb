package filter

import (
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func MessageTypeFromRequest(r *http.Request) (filter.Filter, error) {
	messageTypeQuery, ok := r.URL.Query()["message_type"]
	if !ok || len(messageTypeQuery) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range messageTypeQuery {
		orFilter.Or(
			filter.NewKV("MessageType", v),
		)
	}

	return orFilter, nil
}
