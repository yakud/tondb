package streaming

import (
	"github.com/gobwas/ws/wsutil"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"net"
)

type SubHandler struct {
	Sub *Sub

	Abandoned bool

	resChan   chan []byte
	resBuffer [][]byte

	//some kind of db stuff
	blocksFeed *feed.BlocksFeed

	fromDbCount   uint32
	fetchedFromDb bool
	dbBuffer      [][]byte
}

// TODO: figure out logic of fetching some blocks from db first and then streaming
func (h *SubHandler) Handle() {
	if !h.fetchedFromDb {
		for v := range h.resChan {
			h.resBuffer = append(h.resBuffer, v)
		}
	} else {
		for {
			select {
			case res := <- h.resChan:
				h.resBuffer = append(h.resBuffer, res)
			default:
				// channel is empty, flush buffer to user
				// TODO: we send here separate messages for every entry in buffer, maybe it's better to join all or some
				//  entries with some delimiter and then send them as one message
				for _, res := range h.resBuffer {
					if err := wsutil.WriteServerText(h.Sub.Conn, res); err != nil {
						// TODO: handle error properly, maybe we need to return it?
					}
				}

				h.resBuffer = make([][]byte, 0, 25)
			}
		}
	}
}

func (h *SubHandler) HandleOrAbandon(jsonBytes []byte) {
	select {
	case h.resChan <- jsonBytes:
	default:
		h.Abandoned = true
	}
}


func NewSubHandler(sub *Sub, fromDbCount uint32) SubHandler {
	return SubHandler{
		Sub:           sub,
		Abandoned:     false,
		resChan:       make(chan []byte, 25),
		resBuffer:     make([][]byte, 0, 25),
		fromDbCount:   fromDbCount,
		fetchedFromDb: fromDbCount == 0,
		dbBuffer:      make([][]byte, 0, 8),
	}
}

type ConnSubHandlers struct {
	handlers   []SubHandler
	subManager *SubManager
}

func (h *ConnSubHandlers) AddHandler(conn net.Conn, params Params, id string) *SubHandler {
	var fetchFromDb uint32 = 0
	if params.FetchFromDb != nil {
		fetchFromDb = *params.FetchFromDb
	}

	// sorting because we use type Filter as key in map and we want this key to be correct
	params.CustomFilters.Sort()
	params.Filter.customFilters = params.CustomFilters.String()

	h.handlers = append(h.handlers, NewSubHandler(NewSubUuid(conn, params.Filter, id), fetchFromDb))
	h.subManager.Add(&h.handlers[len(h.handlers)-1])
	return &h.handlers[len(h.handlers)-1]
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
