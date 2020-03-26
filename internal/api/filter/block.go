package filter

import (
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"net/http"
)

func BlockFilterFromRequest(r *http.Request, field string, maxBlocks int) (*filter.Blocks, error) {
	blockQuery, ok := r.URL.Query()[field]
	if !ok || len(blockQuery) == 0 {
		return nil, nil
	}

	blocksId := make([]*ton.BlockId, 0)

	// collect block filters
	for _, vv := range blockQuery {
		block, err := ton.ParseBlockId(vv)
		if err != nil {
			return nil, err
		}
		blocksId = append(blocksId, block)
	}
	if len(blocksId) > maxBlocks {
		return nil, fmt.Errorf("maximum %d blocks per request", maxBlocks)
	}

	return filter.NewBlocks(blocksId...), nil
}

func BlockFilterFromParam(blocks *[]string, maxBlocks int) (*filter.Blocks, error) {
	if blocks == nil {
		return nil, nil
	}

	if len(*blocks) == 0 {
		return nil, nil
	}

	blocksId := make([]*ton.BlockId, 0)

	// collect block filters
	for _, vv := range *blocks {
		block, err := ton.ParseBlockId(vv)
		if err != nil {
			return nil, err
		}
		blocksId = append(blocksId, block)
	}
	if len(blocksId) > maxBlocks {
		return nil, fmt.Errorf("maximum %d blocks per request", maxBlocks)
	}

	return filter.NewBlocks(blocksId...), nil
}