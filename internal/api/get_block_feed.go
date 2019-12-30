package api

import (
	"net/http"
	"strconv"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
)

const defaultBlocksFeedCount = 30

type GetBlocksFeed struct {
	f *feed.BlocksFeed
}

func (m *GetBlocksFeed) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
		limit = defaultBlocksFeedCount
	}

	blocksFeed, err := m.f.SelectBlocks(limit, beforeTime)
	if err != nil {
		http.Error(w, `{"error":true,"message":"error fetch blocks"}`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(blocksFeed)
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
