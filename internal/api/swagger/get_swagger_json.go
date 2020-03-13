package swagger

import (
	"encoding/json"
	"net/http"

	"gitlab.flora.loc/mills/tondb/swagger/tonapi"

	"github.com/julienschmidt/httprouter"
)

type GetSwaggerJson struct {
}

func (m *GetSwaggerJson) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	swagger, err := tonapi.GetSwagger()
	if err != nil {
		http.Error(w, `{"error":true,"message":"get swagger.json error"}`, http.StatusInternalServerError)
		return
	}

	swaggerJson, err := json.Marshal(swagger)
	if err != nil {
		http.Error(w, `{"error":true,"message":"marshal swagger.json error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(swaggerJson)
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
