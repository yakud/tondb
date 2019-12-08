package filter

import (
	"errors"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func TransactionHashByIndexFilterFromRequest(r *http.Request, field string) (*filter.TransactionHashByIndex, error) {
	trHash, ok := r.URL.Query()[field]
	if !ok || len(trHash) == 0 {
		return nil, nil
	}

	if len(trHash) > 1 {
		return nil, errors.New("maximum 1 transaction hash")
	}

	return filter.NewTransactionHashByIndex(trHash[0])
}
