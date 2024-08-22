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
	GetNearestPlaces(lat, lon float64) ([]places.Place, error)
	GetTotalRecords() int
}

func NewStoreAPI(s Store) *API {
	return &API{
		Store: s,
	}
}
