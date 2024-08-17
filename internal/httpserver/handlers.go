package httpserver

import (
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

// я знаю что это не лучшее решение и можно иначе кэшировать страницы
// мне на первых порах кажется что это лучше чем воротить зверинец из
// технологий на 4 дне бассейна и сделать хардкодом чтобы у меня не падал
// по ООМу эластик
var pageCacheMap map[int]PlacesPageContent

const pageSize = 7

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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func validatePageParam(page string, pages int, w http.ResponseWriter) (int, error) {
	if page != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt > pages || pageInt < 1 {
			w.Write([]byte(fmt.Sprintf("Invalid 'page' value: '%v'", page)))
			return 0, err
		}
		return pageInt, nil
	}
	return 0, nil
}
