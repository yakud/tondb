package search

import (
	"log"
	"net/url"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/index"

	"gitlab.flora.loc/mills/tondb/internal/ton/query"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"
)

type Searcher struct {
	accountStorage   *state.AccountState
	getBlockQuery    *query.GetBlockInfo
	indexBlocksSeqNo *index.IndexReverseBlockSeqNo
	indexHash        *index.IndexHash
}

func (s *Searcher) Search(q string) ([]Result, error) {
	q = strings.TrimSpace(q)
	unescapedQuery, err := url.QueryUnescape(q)
	if err == nil {
		q = unescapedQuery
	}

	if result, err := s.searchAccount(q); err == nil {
		return result, nil
	} else {
		log.Println("searchAccount:", err)
	}

	if result, err := s.searchBlockFull(q); err == nil {
		return result, nil
	} else {
		log.Println("searchBlockFull:", err)
	}

	if result, err := s.searchBlocksBySeqNo(q); err == nil {
		return result, nil
	} else {
		log.Println("searchBlocksBySeqNo:", err)
	}

	if result, err := s.searchSomethingByHash(q); err == nil {
		return result, nil
	} else {
		log.Println("searchSomethingByHash:", err)
	}

	return nil, nil
}

func NewSearcher(
	accountStorage *state.AccountState,
	getBlockQuery *query.GetBlockInfo,
	indexBlocksSeqNo *index.IndexReverseBlockSeqNo,
	indexHash *index.IndexHash,
) *Searcher {
	return &Searcher{
		accountStorage:   accountStorage,
		getBlockQuery:    getBlockQuery,
		indexBlocksSeqNo: indexBlocksSeqNo,
		indexHash:        indexHash,
	}
}
