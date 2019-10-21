package filter

import (
	"errors"
	"fmt"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

type BlocksRange struct {
	blockFrom *ton.BlockId
	blockTo   *ton.BlockId
}

func (f *BlocksRange) Build() (string, []interface{}, error) {
	filter := fmt.Sprintf("(WorkchainId = ? AND Shard = ? AND SeqNo >= ? AND SeqNo <= ?)")
	args := []interface{}{
		f.blockFrom.WorkchainId,
		f.blockFrom.Shard,
		f.blockFrom.SeqNo,
		f.blockTo.SeqNo,
	}

	return filter, args, nil
}

func NewBlocksRange(blockFrom *ton.BlockId, blockTo *ton.BlockId) (*BlocksRange, error) {
	if blockFrom.WorkchainId != blockTo.WorkchainId {
		return nil, errors.New("block range filter error different WorkchainId")
	}
	if blockFrom.Shard != blockTo.Shard {
		return nil, errors.New("block range filter error different Shard")
	}

	f := &BlocksRange{
		blockFrom: blockFrom,
		blockTo:   blockTo,
	}

	return f, nil
}
