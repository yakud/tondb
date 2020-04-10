package streaming

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"net"

	"github.com/google/uuid"
)

type Params struct {
	Filter

	FetchFromDb uint32 `json:"fetch_from_db"`
}

type Filter struct {
	FeedName    string `json:"feed_name"`
	WorkchainId *int32  `json:"workchain_id"`
	Shard       *uint64 `json:"shard"`
	AccountAddr *string `json:"account_addr"`
}

type Sub struct {
	Conn   net.Conn
	Filter Filter
	Uuid   string
}

func (f *Filter) Match(block *ton.Block) bool {
	if f.AccountAddr != nil {
		for _, v := range block.Transactions {
			if v.AccountAddr == *f.AccountAddr {
				return true
			}
		}
	}

	return (f.WorkchainId == nil || (f.WorkchainId != nil && block.Info.WorkchainId == *f.WorkchainId)) &&
		(f.Shard == nil || (f.Shard != nil && block.Info.Shard == *f.Shard))
}

func NewSub(conn net.Conn, filter Filter) *Sub {
	return &Sub{
		Conn: conn,
		Filter: filter,
		Uuid: uuid.New().String(),
	}
}

func NewSubUuid(conn net.Conn, filter Filter, id string) *Sub {
	return &Sub{
		Conn: conn,
		Filter: filter,
		Uuid: id,
	}
}



