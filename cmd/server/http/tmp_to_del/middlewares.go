package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func ChainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func ValidatePageParamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "" {
			http.Error(w, "missing page parameter", http.StatusBadRequest)
			return
		}

		if _, err := strconv.Atoi(page); err != nil {
			http.Error(w, "invalid page parameter", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func itemHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	response := map[string]string{"message": "success", "page": page}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
