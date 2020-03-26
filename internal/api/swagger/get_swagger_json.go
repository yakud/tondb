package swagger

import (
	"net/http"

	"gitlab.flora.loc/mills/tondb/swagger/tonapi"

	"github.com/labstack/echo/v4"
)

type GetSwaggerJson struct {
}

func (m *GetSwaggerJson) Handler(ctx echo.Context) error {
	swagger, err := tonapi.GetSwagger()
	if err != nil {
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"get swagger.json error"}`))
	}

	return ctx.JSON(200, swagger)
}

func NewGetSwaggerJson() *GetSwaggerJson {
	return &GetSwaggerJson{}
}

//// swagger doc// (GET /)
//func (api *AuthApi) Get(ctx echo.Context) error {
//	return ctx.Render(http.StatusOK, template.SwaggerDoc, nil)
//}
//
//// swagger.json// (GET /swagger.json)
//func (api *AuthApi) GetSwaggerJson(ctx echo.Context) error {
//	swaggerSpec, err := authapi.GetSwagger()
//	if err != nil {
//		return api.sendError(ctx, http.StatusInternalServerError, "Error fetch swagger json")
//	}
//
//	if err := ctx.JSON(http.StatusOK, swaggerSpec); err != nil {
//		return err
//	}
//
//	return nil
//}
