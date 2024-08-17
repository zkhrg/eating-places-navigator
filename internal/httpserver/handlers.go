package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/zkhrg/go_day03/internal/elasticsearch"
)

type PlacesPageContent struct {
	Total    int
	Items    []elasticsearch.PlacesHit
	Page     int
	NextPage int
	PrevPage int
	LastPage int
}

type PlacesAPISchema struct {
	Name     string        `json:"name"`
	Total    int           `json:"total"`
	Places   []PlaceSchema `json:"places"`
	PrevPage int           `json:"prev_page"`
	NextPage int           `json:"next_page"`
	LastPage int           `json:"last_page"`
}

type PlaceSchema struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"location"`
}

// я знаю что это не лучшее решение и можно иначе кэшировать страницы
// мне на первых порах кажется что это лучше чем воротить зверинец из
// технологий на 4 дне бассейна и сделать хардкодом чтобы у меня не падал
// по ООМу эластик
var pageCacheMap map[int]PlacesPageContent

const pageSize = 11

// StartPageHandler обрабатывает только GET-запросы на /
func StartPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
		return
	}
	var pgc PlacesPageContent
	total := elasticsearch.CountIndexRecords("places")
	pages := total / pageSize
	if total%pageSize != 0 {
		pages += 1
	}
	if len(pageCacheMap) == 0 {
		pageCacheMap = make(map[int]PlacesPageContent)
	}
	tmpl, _ := template.ParseFiles("templates/index.html")

	page, err := validatePageParam(r.URL.Query().Get("page"), pages, w)
	if err != nil {
		return
	}

	if val, ok := pageCacheMap[page]; ok {
		tmpl.Execute(w, val)
		return
	}
	items := elasticsearch.GetPageData(page, pageSize, "places")

	pgc = PlacesPageContent{
		Total:    total,
		Items:    items,
		Page:     page,
		PrevPage: page - 1,
		NextPage: page + 1,
		LastPage: pages,
	}

	if err := tmpl.Execute(w, pgc); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Template execution error:", err)
		return
	}
	pageCacheMap[page] = pgc
}

func SimpleAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
		return
	}

	total := elasticsearch.CountIndexRecords("places")
	pages := total / pageSize
	if total%pageSize != 0 {
		pages += 1
	}

	page, err := validatePageParam(r.URL.Query().Get("page"), pages, w)
	if err != nil {
		return
	}
	pas := PlacesAPISchema{
		Total:    total,
		LastPage: pages,
		PrevPage: page - 1,
		NextPage: page + 1,
		Name:     "palces",
		Places:   make([]PlaceSchema, 0),
	}
	data := elasticsearch.GetPageData(page, pageSize, "places")
	for _, v := range data {
		id, _ := strconv.Atoi(v.Source.ID)
		ps := PlaceSchema{
			ID:      id,
			Name:    v.Source.Name,
			Address: v.Source.Address,
			Phone:   v.Source.Phone,
		}
		lat, _ := strconv.ParseFloat(v.Source.Location.Lat, 64)
		lon, _ := strconv.ParseFloat(v.Source.Location.Lon, 64)
		ps.Location.Lat = lat
		ps.Location.Lon = lon
		pas.Places = append(pas.Places, ps)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pas)
}

func validatePageParam(page string, pages int, w http.ResponseWriter) (int, error) {
	if page != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt > pages || pageInt < 1 {
			w.Write([]byte(fmt.Sprintf("Invalid 'page' value: '%v'", page)))
			fmt.Println("as")
			return 0, errors.New("invalid page value")
		}
		return pageInt, nil
	}
	w.Write([]byte("no page argument provided"))
	return 0, errors.New("no page argument provided")
}
