package search

import (
	"fmt"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/index"
)

func (s *Searcher) searchSomethingByHash(q string) ([]Result, error) {
	if len(q) != 64 {
		return nil, fmt.Errorf("is not hash")
	}

	something, err := s.indexHash.SelectSomethingByHash(q)
	if err != nil {
		return nil, fmt.Errorf("error select something by hash: %w", err)
	}

	var result []Result
	for _, s := range something {
		var searchResultType ResultType
		var link string

		switch s.Type {
		case index.TypeBlock:
			searchResultType = ResultTypeBlock
			link = "/block/info?block=" + strings.ToUpper(s.Data)

		case index.TypeTransaction:
			searchResultType = ResultTypeTransaction
			link = "/transaction?hash=" + strings.ToUpper(s.Hash)

			// TODO: add message

		default:
			continue
		}

		result = append(result, Result{
			Type: searchResultType,
			Hint: q,
			Link: link,
		})
	}

	return result, nil
}
