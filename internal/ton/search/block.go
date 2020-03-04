package search

import (
	"fmt"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func (s *Searcher) searchBlockFull(q string) ([]Result, error) {
	blockId, err := ton.ParseBlockId(q)
	if err != nil {
		return nil, fmt.Errorf("error parse blocks id full '%s', %w", q, err)
	}

	blocks, err := s.getBlockQuery.GetBlockInfo(filter.NewBlocks(blockId))
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("block not found '%s'", q)
	}

	var result []Result
	for _, block := range blocks {
		blockIdStr := block.BlockId.String()
		result = append(result, Result{
			Type: ResultTypeBlock,
			Hint: blockIdStr,
			Link: "/block/info?block=" + blockIdStr,
		})
	}

	return result, nil
}

func (s *Searcher) searchBlocksBySeqNo(q string) ([]Result, error) {
	blockSeqNo, err := strconv.ParseUint(q, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("is not number '%s': %w", q, err)
	}

	blocksId, err := s.indexBlocksSeqNo.SelectBlocksBySeqNo(blockSeqNo)
	if err != nil {
		return nil, err
	}

	blocks, err := s.getBlockQuery.GetBlockInfo(filter.NewBlocks(blocksId...))
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("block not found")
	}

	var result []Result
	for _, block := range blocks {
		blockIdStr := block.BlockId.String()
		result = append(result, Result{
			Type: ResultTypeBlock,
			Hint: blockIdStr,
			Link: "/block/info?block=" + blockIdStr,
		})
	}

	return result, nil
}

func (s *Searcher) isNumber(query string) bool {
	if _, err := strconv.ParseUint(query, 10, 64); err == nil {
		return true
	}

	return false
}

func (s *Searcher) isFullBlockNum(query string) bool {
	if _, err := ton.ParseBlockId(query); err == nil {
		return true
	}

	return false
}
