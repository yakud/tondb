package streaming

import (
	"github.com/gobwas/ws/wsutil"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"net"
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
		return wsutil.WriteServerText(h.Sub.Conn, res)
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

type ConnSubHandlers struct {
	handlers   []SubHandler
	subManager *SubManager
}

func (h *ConnSubHandlers) AddHandler(conn net.Conn, params Params, id string) {
	h.handlers = append(h.handlers, NewSubHandler(NewSubUuid(conn, params.Filter, id), params.FetchFromDb))
	h.subManager.Add(&h.handlers[len(h.handlers)-1])
}

func (h *ConnSubHandlers) RemoveHandler(id string) {
	for i := range h.handlers {
		if h.handlers[i].Sub.Uuid == id {
			h.handlers[i].Abandoned = true
			h.handlers = append(h.handlers[:i], h.handlers[i+1:]...)
		}
	}
}

func NewConnSubHandlers(manager *SubManager) *ConnSubHandlers {
	return &ConnSubHandlers{
		handlers:   make([]SubHandler, 0, 4),
		subManager: manager,
	}
}
