package site

import (
	"encoding/json"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
)

const defaultLatestMessagesCount = 500

type GetLatestMessages struct {
	q *feed.MessagesFeedGlobal
}

func (api *GetLatestMessages) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	latestMessages, err := api.q.SelectLatestMessages(defaultLatestMessagesCount)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieve last messages from DB"}`))
		return
	}

	resp, err := json.Marshal(latestMessages)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serialize response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewGetLatestMessages(q *feed.MessagesFeedGlobal) *GetLatestMessages {
	return &GetLatestMessages{
		q: q,
	}
}
