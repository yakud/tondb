package search

import (
	"fmt"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/index"
)

func (s *Searcher) searchSomethingByHash(q string) ([]tonapi.SearchResult, error) {
	if len(q) != 64 {
		return nil, fmt.Errorf("is not hash")
	}

	q = strings.ToUpper(q)

	something, err := s.indexHash.SelectSomethingByHash(q)
	if err != nil {
		return nil, fmt.Errorf("error select something by hash: %w", err)
	}

	var result []tonapi.SearchResult
	for _, s := range something {
		var searchResultType ResultType
		var link string
		var hint string

		switch s.Type {
		case index.TypeBlock:
			searchResultType = ResultTypeBlock
			link = "/block/info?block=" + strings.ToUpper(s.Data)
			hint = s.Data

		case index.TypeTransaction:
			searchResultType = ResultTypeTransaction
			link = "/transaction?hash=" + strings.ToUpper(s.Hash)
			hint = q

			// TODO: add message

		default:
			continue
		}

		result = append(result, tonapi.SearchResult{
			Type: string(searchResultType),
			Hint: hint,
			Link: link,
		})
	}

	return result, nil
}
