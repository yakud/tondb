package api

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
	filter2 "gitlab.flora.loc/mills/tondb/internal/api/filter"
)

const defaultTransactionsCount = 30

type GetAccountTransactions struct {
	f *feed.AccountTransactions
}

func (m *GetAccountTransactions) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// address
	accountFilter, err := filter2.AccountFilterFromRequest(r, "address")
	if err != nil {
		http.Error(w, `{"error":true,"message":"error make account filter: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// before_lt
	var beforeLt uint64
	beforeLtStr, ok := r.URL.Query()["before_lt"]
	if ok {
		if len(beforeLtStr) > 1 {
			http.Error(w, `{"error":true,"message":"should be set only one before_lt field"}`, http.StatusBadRequest)
			return
		}
		beforeLt, err = strconv.ParseUint(beforeLtStr[0], 10, 64)
		if err != nil {
			http.Error(w, `{"error":true,"message":"error parsing before_lt field"}`, http.StatusBadRequest)
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
		limit = defaultTransactionsCount
	}

	accountTransactions, err := m.f.GetAccountTransactions(accountFilter.Addr(), beforeLt, limit, nil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, `{"error":true,"message":"error fetch transactions"}`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(accountTransactions)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetAccountTransactions(f *feed.AccountTransactions) *GetAccountTransactions {
	return &GetAccountTransactions{
		f: f,
	}
}
