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

// HelloHandler обрабатывает только GET-запросы на /hello
func StartPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
		return
	}
	var pgc PlacesPageContent

	tmpl, _ := template.ParseFiles("templates/index.html")

	str_page := r.URL.Query().Get("page")
	if str_page != "" {
		page, err := strconv.Atoi(str_page)
		if err != nil {
			return
		}
		pageSize := 10
		items, total := elasticsearch.GetPageData(page, pageSize, "places")
		pages := total / pageSize
		if total%pageSize != 0 {
			pages += 1
		}
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
	}
}
