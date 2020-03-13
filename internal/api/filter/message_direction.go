package filter

import (
	"fmt"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func MessageDirectionFromRequest(r *http.Request) (filter.Filter, error) {
	dirQuery, ok := r.URL.Query()["dir"]
	if !ok || len(dirQuery) == 0 {
		return nil, nil
	}

	switch dirQuery[0] {
	case "in":
		return filter.NewArrayHas("Messages.Direction", "in"), nil
	case "out":
		return filter.NewArrayHas("Messages.Direction", "out"), nil
	default:
		return nil, fmt.Errorf("undefined dir value: %s", dirQuery[0])
	}
}
