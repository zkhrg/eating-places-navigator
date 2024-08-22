package http

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"

	"github.com/zkhrg/go_day03/internal/api"
)

func HTMLPageHandler(a *api.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := a.GetPage(r.Context(), r.Context().Value(PageContextKey).(int), 10)
		if err != nil {
			http.Error(w, "Failed to get page", http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("cmd/server/http/web/templates/index.html")
		if err != nil {
			http.Error(w, "Template parsing error", http.StatusInternalServerError)
			log.Println("Error parsing template:", err)
			return
		}

		err = tmpl.Execute(w, page)
		if err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
			log.Println("Error executing template:", err)
			return
		}
	}
}

func JSONPageHandler(a *api.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := a.GetPage(r.Context(), r.Context().Value(PageContextKey).(int), 10)
		if err != nil {
			http.Error(w, "Failed to get page", http.StatusInternalServerError)
			return
		}
		// Устанавливаем заголовок Content-Type
		w.Header().Set("Content-Type", "application/json")
		page.Name = "places"
		// Сериализуем данные в формат JSON
		jsonResponse, err := json.Marshal(page)
		if err != nil {
			http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
			return
		}

		// Отправляем JSON-ответ
		w.Write(jsonResponse)
	}
}
