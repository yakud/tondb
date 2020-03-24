package timeseries

import (
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"

	"github.com/labstack/echo/v4"
)

type VolumeByGrams struct {
	q *timeseries.VolumeByGrams
}

func (api *VolumeByGrams) GetV1TimeseriesVolumeByGrams(ctx echo.Context) error {
	res, err := api.q.GetVolumeByGrams()
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving timeseries"}`))
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewVolumeByGrams(q *timeseries.VolumeByGrams) *VolumeByGrams {
	return &VolumeByGrams{
		q: q,
	}
}
