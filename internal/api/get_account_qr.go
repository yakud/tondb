package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	apifilter "gitlab.flora.loc/mills/tondb/internal/api/filter"
	"gitlab.flora.loc/mills/tondb/internal/utils"

	"github.com/julienschmidt/httprouter"
	"github.com/skip2/go-qrcode"
)

const qrFormatPNG = "png"
const qrFormatPNGBase64 = "png_base64"

type GetAccountQR struct {
}

func (m *GetAccountQR) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	accountFilter, err := apifilter.AccountFilterFromRequest(r, "address")
	if err != nil {
		http.Error(w, `{"error":true,"message":"error make account filter: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	var imageFormat string
	if format, ok := r.URL.Query()["format"]; !ok || len(format) == 0 {
		imageFormat = qrFormatPNG
	} else {
		imageFormat = format[0]
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

	switch imageFormat {
	case qrFormatPNGBase64:
		w.Header().Add("Content-Type", "text/plain")
		png = []byte(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(png)))

	case qrFormatPNG:
		w.Header().Add("Content-Type", "image/png")

	default:
		http.Error(w, `{"error":true,"message":"error wrong format string"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(200)
	w.Write(png)
}

func NewGetAccountQR() *GetAccountQR {
	return &GetAccountQR{}
}
