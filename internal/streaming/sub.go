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
	WorkchainId int32  `json:"workchain_id"`
	Shard       uint64 `json:"shard"`
	AccountAddr string `json:"account_addr"`
}

type Sub struct {
	Conn   net.Conn
	Filter Filter
	Uuid   string
}

func (f *Filter) Match(block *ton.Block) (ok bool) {
	for _, v := range block.Transactions {
		ok = v.AccountAddr == f.AccountAddr
	}

	// TODO: rewrite this logic, it is not how it should be)
	return block.Info.WorkchainId == f.WorkchainId || block.Info.Shard == f.Shard || ok
}

func NewSub(conn net.Conn, filter Filter) *Sub {
	return &Sub{
		Conn: conn,
		Filter: filter,
		Uuid: uuid.New().String(),
	}
}




