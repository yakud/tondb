package filter

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func BlockRangeFilterFromRequest(r *http.Request, fieldFrom, fieldTo string, maxRange int) (*filter.BlocksRange, error) {
	blockFromQuery, ok := r.URL.Query()[fieldFrom]
	if !ok || len(blockFromQuery) == 0 {
		return nil, nil
	}

	blockToQuery, ok := r.URL.Query()[fieldTo]
	if !ok || len(blockToQuery) == 0 {
		return nil, nil
	}

	blockFrom, err := ton.ParseBlockId(blockFromQuery[0])
	if err != nil {
		return nil, err
	}
	blockTo, err := ton.ParseBlockId(blockToQuery[0])
	if err != nil {
		return nil, err
	}

	if blockTo.SeqNo < blockFrom.SeqNo {
		return nil, errors.New("block_from should be less or equals then block_to")
	}

	if blockTo.SeqNo-blockFrom.SeqNo > uint64(maxRange) {
		return nil, fmt.Errorf("maximum %d blocks per request", uint64(maxRange))
	}

	filterBlockRange, err := filter.NewBlocksRange(blockFrom, blockTo)
	if err != nil {
		return nil, err
	}

	return filterBlockRange, nil
}
