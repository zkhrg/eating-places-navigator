package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/zkhrg/go_day03/internal/api"
)

type contextKey string

const PageContextKey contextKey = "page"

func ChainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Проходим по всем миддлварям в обратном порядке, чтобы
	// первый миддлварь был самым внешним, а последний - самым внутренним
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func PaginationMiddleware(a *api.API) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем значение параметра `page`
			pageParam := r.URL.Query().Get("page")

			// Проверяем наличие параметра `page`
			if pageParam == "" {
				http.Error(w, "Missing 'page' parameter", http.StatusBadRequest)
				return
			}

			// Преобразуем параметр `page` в int
			page, err := strconv.Atoi(pageParam)
			if err != nil || page < 1 || page > api.GetPagesCount(10, a.Store.GetTotalRecords()) {
				http.Error(w, "'page' parameter must be a positive integer and dont overflow pages count", http.StatusBadRequest)
				return
			}

			// Передаем параметр `page` в контексте
			ctx := context.WithValue(r.Context(), PageContextKey, page)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
