package api

import (
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"

	filter2 "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"github.com/julienschmidt/httprouter"
)

type GetAccount struct {
	s *state.AccountState
}

func (m *GetAccount) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	accountFilter, err := filter2.AccountFilterFromRequest(r, "address")
	if err != nil {
		http.Error(w, `{"error":true,"message":"error make account filter: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	accountState, err := m.s.GetAccount(accountFilter)
	if err != nil {
		http.Error(w, `{"error":true,"message":"error fetch account: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(accountState)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetAccount(s *state.AccountState) *GetAccount {
	return &GetAccount{
		s: s,
	}
}
