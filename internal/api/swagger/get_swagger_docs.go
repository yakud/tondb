package swagger

import (
	"github.com/labstack/echo/v4"
)

type GetSwaggerDocs struct {
}

func (m *GetSwaggerDocs) Handler(ctx echo.Context) error {
	return ctx.HTML(200, redocHtml)
}

func NewGetSwaggerDocs() *GetSwaggerDocs {
	return &GetSwaggerDocs{}
}

const redocHtml = `<!DOCTYPE html>
<html>

<head>
    <meta charset="UTF-8" />
    <title>Auth-API Documentation</title>
    <style>
        body {
            margin: 0;
            padding: 0;
        }

        redoc {
            display: block;
        }
    </style>
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
</head>

<body>
<redoc spec-url="/swagger.json"></redoc>
<script src="https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"></script>
</body>

</html>`
