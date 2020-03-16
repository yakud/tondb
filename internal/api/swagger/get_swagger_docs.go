package swagger

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type GetSwaggerDocs struct {
}

func (m *GetSwaggerDocs) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write(redocHtml)
}

func NewGetSwaggerDocs() *GetSwaggerDocs {
	return &GetSwaggerDocs{}
}

var redocHtml = []byte(`<!DOCTYPE html>
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

</html>`)
