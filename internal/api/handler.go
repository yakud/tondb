package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handler interface {
	Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}
