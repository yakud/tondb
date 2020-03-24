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
			filter.NewArrayHas("Messages.Type", v),
		)
	}

	return orFilter, nil
}

func MessageTypeFromParam(messageType *[]string) (filter.Filter, error) {
	if messageType == nil || len(*messageType) == 0 {
		return nil, nil
	}

	orFilter := filter.NewOr()
	for _, v := range *messageType {
		orFilter.Or(
			filter.NewArrayHas("Messages.Type", v),
		)
	}

	return orFilter, nil
}

