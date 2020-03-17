package main

import (
	"fmt"
	"os"

	"gitlab.flora.loc/mills/tondb/internal/server"

	"gitlab.flora.loc/mills/tondb/swagger/tonapi"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	swagger, err := tonapi.GetSwagger()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error loading swagger spec\n: %s", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	tonApiServer := &server.TonApi{} // todo: constructor func
	tonapi.RegisterHandlers(e, tonApiServer)

	if err := e.Start("0.0.0.0:112233"); err != nil {
		e.Logger.Fatal(err)
	}
}
