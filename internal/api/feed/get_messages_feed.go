package feed

import (
	"encoding/json"
	"log"
	"net/http"

	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
)

const defaultLatestMessagesCount = 50

type GetMessagesFeedResponse struct {
	Messages []*feed.MessageInFeed `json:"messages"`
	ScrollId string                `json:"scroll_id"`
}

type GetMessagesFeed struct {
	q *feed.MessagesFeed
}

func (api *GetMessagesFeed) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error

	// limit
	limit, err := httputils.GetQueryValueUint16(r.URL, "limit")
	if err != nil {
		limit = defaultLatestMessagesCount
	}

	// workchain_id
	workchainId, err := httputils.GetQueryValueInt32(r.URL, "workchain_id")
	if err != nil {
		workchainId = feed.EmptyWorkchainId
	}

	// scroll_id
	var scrollId = &feed.MessagesFeedScrollId{}
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

	messagesFeed, newScrollId, err := api.q.SelectMessages(scrollId, limit)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error retrieve messages from DB"}`))
		return
	}
	newPackedScrollId, err := PackScrollId(newScrollId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error pack scroll_id"}`))
		return
	}

	resp := GetMessagesFeedResponse{
		Messages: messagesFeed,
		ScrollId: newPackedScrollId,
	}

	respJson, err := json.Marshal(&resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error serialize response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func NewGetMessagesFeed(q *feed.MessagesFeed) *GetMessagesFeed {
	return &GetMessagesFeed{
		q: q,
	}
}
