package feed

import (
	"encoding/json"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"

	"github.com/julienschmidt/httprouter"
)

const (
	defaultBlocksFeedCount = 30
)

type GetBlocksFeedResponse struct {
	Blocks   []*feed.BlockInFeed `json:"blocks"`
	ScrollId string              `json:"scroll_id"`
}

type GetBlocksFeed struct {
	f *feed.BlocksFeed
}

func (m *GetBlocksFeed) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error

	// limit
	limit, err := httputils.GetQueryValueUint16(r.URL, "limit")
	if err != nil {
		limit = defaultBlocksFeedCount
	}

	// workchain_id
	workchainId, err := httputils.GetQueryValueInt32(r.URL, "workchain_id")
	if err != nil {
		workchainId = feed.EmptyWorkchainId
	}

	// scroll_id
	var scrollId = &feed.BlocksFeedScrollId{}
	packedScrollId, err := httputils.GetQueryValueString(r.URL, "scroll_id")
	if err == nil && len(packedScrollId) > 0 {
		if err := UnpackScrollId(packedScrollId, scrollId); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":true,"message":"error unpack scroll_id"}`))
			return
		}
	} else {
		scrollId.WorkchainId = workchainId
	}

	blocksFeed, newScrollId, err := m.f.SelectBlocks(scrollId, limit)
	if err != nil {
		http.Error(w, `{"error":true,"message":"error fetch blocks"}`, http.StatusInternalServerError)
		return
	}

	newPackedScrollId, err := PackScrollId(newScrollId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error pack scroll_id"}`))
		return
	}

	respFeed := GetBlocksFeedResponse{
		Blocks:   blocksFeed,
		ScrollId: newPackedScrollId,
	}

	resp, err := json.Marshal(&respFeed)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetBlocksFeed(f *feed.BlocksFeed) *GetBlocksFeed {
	return &GetBlocksFeed{
		f: f,
	}
}
