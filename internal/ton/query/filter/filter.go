package filter

import (
	"fmt"
)

type Filter interface {
	Build() (sqlFilter string, args []interface{}, err error)
}

type Filters []Filter

func (f *Filters) Add(filter ...Filter) {
	if filter != nil && len(filter) > 0 {
		*f = append(*f, filter...)
	}
}

// Render filters and include in query
// Expected query like: SELECT 1 FROM table WHERE %s
func RenderQuery(query string, filter Filter) (string, []interface{}, error) {
	var queryWhere string
	var err error
	var filtersArgs = make([]interface{}, 0)

	if filter == nil {
		queryWhere = "1=1"
	} else {
		queryWhere, filtersArgs, err = filter.Build()
		if err != nil {
			return "", nil, err
		}
	}
	return fmt.Sprintf(query, queryWhere), filtersArgs, nil
}
