package api

import (
	"encoding/base64"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/utils"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
)

const qrFormatPNG = "png"
const qrFormatPNGBase64 = "png_base64"

func (s *TonApiServer) GetV1AccountQr(ctx echo.Context, params tonapi.GetV1AccountQrParams) error {
	accAddr, err := ton.ParseAccountAddress(strings.TrimSpace(params.Address))
	if err != nil {
		return err
	}

	accountFilter := filter.NewAccount(accAddr)

	var imageFormat string
	if params.Format == nil || len(*params.Format) == 0 {
		imageFormat = qrFormatPNG
	} else {
		imageFormat = *params.Format
	}

	ufAddr, err := utils.ComposeRawAndConvertToUserFriendly(accountFilter.Addr().WorkchainId, accountFilter.Addr().Addr)
	if err != nil {
		return err
	}

	link := "ton://transfer/" + base64.RawURLEncoding.EncodeToString([]byte(ufAddr))

	var png []byte
	png, err = qrcode.Encode(link, qrcode.Medium, 256)
	if err != nil {
		return err
	}

	responseWriter := ctx.Response().Writer

	switch imageFormat {
	case qrFormatPNGBase64:
		responseWriter.Header().Add("Content-Type", "text/plain")
		png = []byte(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(png)))

	case qrFormatPNG:
		responseWriter.Header().Add("Content-Type", "image/png")

	default:
		return err
	}

	responseWriter.WriteHeader(200)
	_, err = responseWriter.Write(png)

	return err
}