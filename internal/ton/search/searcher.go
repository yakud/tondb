package search

/*
type Searcher struct {
}

func (s *Searcher) Search(query string) ([]Result, error) {
	if s.isAccountAddress(query) {

	}
	if s.isFullBlockNum(query) {

	}
}

func (s *Searcher) isAccountAddress(query string) bool {
	if strings.Contains(query, ":") {
		if parts := strings.Split(query, ":"); len(parts) == 2 && len(parts[1]) == 64 {
			return true
		}
	}
	if len(query) == 48 {
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

func NewSearcher() *Searcher {
	return &Searcher{}
}
*/
