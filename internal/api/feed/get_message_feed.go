package feed

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
)

const defaultLatestMessagesCount = 50

type GetMessagesFeed struct {
	q *feed.MessagesFeed
}

func (api *GetMessagesFeed) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error

	// before_time
	var beforeTime time.Time
	beforeTimeStr, ok := r.URL.Query()["before_time"]
	if ok {
		if len(beforeTimeStr) > 1 {
			http.Error(w, `{"error":true,"message":"should be set only one before_time field"}`, http.StatusBadRequest)
			return
		}
		beforeTimeInt, err := strconv.ParseInt(beforeTimeStr[0], 10, 64)
		if err != nil {
			http.Error(w, `{"error":true,"message":"error parsing before_time field"}`, http.StatusBadRequest)
			return
		}

		beforeTime = time.Unix(beforeTimeInt, 0).UTC()
	} else {
		beforeTime = time.Time{}
	}

	// limit
	var limit int16
	limitStr, ok := r.URL.Query()["limit"]
	if ok {
		if len(limitStr) > 1 {
			http.Error(w, `{"error":true,"message":"should be set only one limit field"}`, http.StatusBadRequest)
			return
		}
		limit64, err := strconv.ParseInt(limitStr[0], 10, 16)
		if err != nil {
			http.Error(w, `{"error":true,"message":"error parsing limit field"}`, http.StatusBadRequest)
			return
		}
		limit = int16(limit64)
	} else {
		limit = defaultLatestMessagesCount
	}

	// workchain_id
	var wcId int32
	wcIdStr, ok := r.URL.Query()["workchain_id"]
	if ok {
		if len(wcIdStr) > 1 {
			http.Error(w, `{"error":true,"message":"only one workchain_id parameter can be set"}`, http.StatusBadRequest)
			return
		}
		wcId64, err := strconv.ParseInt(wcIdStr[0], 10, 32)
		if err != nil {
			http.Error(w, `{"error":true,"message":"error parsing workchain_id field"}`, http.StatusBadRequest)
			return
		}
		wcId = int32(wcId64)
	} else {
		wcId = noWcId
	}

	messagesFeed, err := api.q.SelectLatestMessages(wcId, limit, beforeTime)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"error retrieve messages from DB"}`))
		return
	}

	resp, err := json.Marshal(messagesFeed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"error serialize response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewGetMessagesFeed(q *feed.MessagesFeed) *GetMessagesFeed {
	return &GetMessagesFeed{
		q: q,
	}
}
