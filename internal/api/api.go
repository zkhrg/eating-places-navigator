package api

import (
	"context"

	"github.com/zkhrg/go_day03/internal/places"
)

type API struct {
	Store Store
}

type Store interface {
	GetPlacesByPageParams(ctx context.Context, pageNumber int, pageSize int) ([]places.Place, error)
	GetTotalRecords() int
}

func New(s Store) *API {
	return &API{
		Store: s,
	}
}
