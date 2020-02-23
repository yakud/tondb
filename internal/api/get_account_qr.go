package api

import (
	"encoding/base64"
	"net/http"

	apifilter "gitlab.flora.loc/mills/tondb/internal/api/filter"
	"gitlab.flora.loc/mills/tondb/internal/utils"

	"github.com/julienschmidt/httprouter"
	"github.com/skip2/go-qrcode"
)

type GetAccountQR struct {
}

func (m *GetAccountQR) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	accountFilter, err := apifilter.AccountFilterFromRequest(r, "address")
	if err != nil {
		http.Error(w, `{"error":true,"message":"error make account filter: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	ufAddr, err := utils.ComposeRawAndConvertToUserFriendly(accountFilter.Addr().WorkchainId, accountFilter.Addr().Addr)
	if err != nil {
		http.Error(w, `{"error":true,"message":"error address convertation"}`, http.StatusInternalServerError)
		return
	}

	link := "ton://transfer/" + base64.RawURLEncoding.EncodeToString([]byte(ufAddr))

	var png []byte
	png, err = qrcode.Encode(link, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, `{"error":true,"message":"error QR code generation"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "image/png")
	w.WriteHeader(200)
	w.Write(png)
}

func NewGetAccountQR() *GetAccountQR {
	return &GetAccountQR{}
}
