package http

import (
	"errors"
	"net/http"
	"text/template"
)

type Handlers struct {
	apis api.Server
	home *template.Template
}

type Store interface {
	GetPlaces(pageNumber, pageSize int) ([]places.PlaceModel, errr)
}

// не понял как это имплементировать на стандартный пакет,
// так что трогать не буду пока
// func (h *Handlers) routes() {}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) error {
	out, err := h.apis.ServerHealth()
	if err != nil {
		return err
	}
	// r200 + writer + out of 'out'
	// webgo.R200(w, out)
	return nil
}

func errWrapper(h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			return
		}

		status, msg, _ := errors.HTTPStatusCodeMessage(err)
	}
}
