package filter

import (
	"fmt"
	"strings"
)

type SrcDestAddr struct {
	addr []string
}

func (f *SrcDestAddr) Build() (string, []interface{}, error) {
	filters := make([]string, 0, len(f.addr))
	args := make([]interface{}, 0, len(f.addr)*3)
	for _, a := range f.addr {
		filters = append(filters, "(MessageDestAddr = ? OR MessageSrcAddr = ?)")
		args = append(args, a, a)
	}

	filter := fmt.Sprintf(
		"(%s)",
		strings.Join(filters, "OR"),
	)

	return filter, args, nil
}

func NewSrcDestAddr(addr ...string) *SrcDestAddr {
	f := &SrcDestAddr{
		addr: make([]string, len(addr)),
	}

	copy(f.addr, addr)

	return f
}
