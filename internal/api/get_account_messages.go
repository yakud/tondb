package api

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
	feed2 "gitlab.flora.loc/mills/tondb/internal/api/feed"
	filter2 "gitlab.flora.loc/mills/tondb/internal/api/filter"
	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"
)

const defaultMessagesCount = 30

type GetAccountMessages struct {
	f *feed.AccountMessages
}

type GetAccountMessagesResponse struct {
	Messages []*feed.AccountMessage `json:"messages"`
	ScrollId string                 `json:"scroll_id"`
}

func (m *GetAccountMessages) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// address
	accountFilter, err := filter2.AccountFilterFromRequest(r, "address")
	if err != nil {
		http.Error(w, `{"error":true,"message":"error make account filter: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// scroll_id
	var scrollId = &feed.AccountMessagesScrollId{}
	packedScrollId, err := httputils.GetQueryValueString(r.URL, "scroll_id")
	if err == nil && len(packedScrollId) > 0 {
		if err := feed2.UnpackScrollId(packedScrollId, scrollId); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":true,"message":"error unpack scroll_id"}`))
			return
		}
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
		limit = defaultMessagesCount
	}

	accountMessages, newScrollId, err := m.f.GetAccountMessages(accountFilter.Addr(), scrollId, limit, nil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, `{"error":true,"message":"error fetching messages"}`, http.StatusInternalServerError)
		return
	}

	newPackedScrollId, err := feed2.PackScrollId(newScrollId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error packing scroll_id"}`))
		return
	}

	messagesResponse := GetAccountMessagesResponse{
		Messages: accountMessages,
		ScrollId: newPackedScrollId,
	}

	resp, err := json.Marshal(messagesResponse)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetAccountMessages(f *feed.AccountMessages) *GetAccountMessages {
	return &GetAccountMessages{
		f: f,
	}
}
