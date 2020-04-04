package streaming

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

type SubHandler struct {
	Sub *Sub

	Abandoned bool

	//some kind of db stuff
	blocksFeed *feed.BlocksFeed

	fromDbCount   uint32
	fetchedFromDb bool
	resBuffer     [][]byte
}

// TODO: figure out logic of fetching some blocks from db first and then streaming

func (h *SubHandler) Handle(res []byte) error {
	if !h.fetchedFromDb {
		h.resBuffer = append(h.resBuffer, res)
		return nil
	} else {
		return wsutil.WriteServerMessage(h.Sub.Conn, ws.OpText, res)
	}
}

func NewSubHandler(sub *Sub, fromDbCount uint32) SubHandler {
	return SubHandler{
		Sub:           sub,
		Abandoned:     false,
		fromDbCount:   fromDbCount,
		fetchedFromDb: fromDbCount == 0,
		resBuffer:     make([][]byte, 0, 8),
	}
}
