package filter

import (
	"fmt"
	"strings"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton"
)

type Blocks struct {
	blocks []*ton.BlockId
}

func (f *Blocks) Build() (string, []interface{}, error) {
	filters := make([]string, 0, len(f.blocks))
	args := make([]interface{}, 0, len(f.blocks)*3)
	for _, b := range f.blocks {
		filters = append(filters, "(?,?,?)")
		args = append(args, b.WorkchainId, b.Shard, b.SeqNo)
	}

	filter := fmt.Sprintf(
		"((WorkchainId, Shard, SeqNo) IN (%s))",
		strings.Join(filters, ","),
	)

	return filter, args, nil
}

func NewBlocks(blocks ...*ton.BlockId) *Blocks {
	f := &Blocks{
		blocks: make([]*ton.BlockId, len(blocks)),
	}

	copy(f.blocks, blocks)

	return f
}
