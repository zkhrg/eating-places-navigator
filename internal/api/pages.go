package api

import (
	"context"
	"fmt"

	"github.com/zkhrg/go_day03/internal/places"
)

type Page struct {
	Name     string         `json:"name"`
	Total    int            `json:"total"`
	Places   []places.Place `json:"places"`
	PrevPage int            `json:"prev_page"`
	NextPage int            `json:"next_page"`
	LastPage int            `json:"last_page"`
}

func (a *API) GetPage(ctx context.Context, pageNumber int, pageSize int) (Page, error) {
	// дергает метод из эластика и просто его возвращает
	places, err := a.Store.GetPlacesByPageParams(ctx, pageNumber, pageSize)
	if err != nil {
		fmt.Println("hanling error at getpage")
	}
	total := a.Store.GetTotalRecords()
	return Page{
		Places: places,
		Total:  total,
		// PrevPage: ,
	}, nil
}
