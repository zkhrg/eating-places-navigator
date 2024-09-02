package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/zkhrg/go_day03/internal/api"
)

type contextKey string

const (
	PageContextKey     contextKey = "page"
	LatContextKey      contextKey = "lat"
	LonContextKey      contextKey = "lon"
	UsernameContextKey contextKey = "username"
)

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
			pageParam := r.URL.Query().Get("page")

			if pageParam == "" {
				http.Error(w, "Missing 'page' parameter", http.StatusBadRequest)
				return
			}

			page, err := strconv.Atoi(pageParam)
			if err != nil || page < 1 || page > api.GetPagesCount(10, a.Store.GetTotalRecords()) {
				http.Error(w, "'page' parameter must be a positive integer and dont overflow pages count", http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), PageContextKey, page)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LatLonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		latParam := r.URL.Query().Get("lat")
		lonParam := r.URL.Query().Get("lon")

		if latParam == "" || lonParam == "" {
			http.Error(w, "Missing 'lat' or 'lon' parameter", http.StatusBadRequest)
			return
		}

		lat, err := strconv.ParseFloat(latParam, 64)
		if err != nil {
			http.Error(w, "'lat' parameter must be a valid float", http.StatusBadRequest)
			return
		}

		lon, err := strconv.ParseFloat(lonParam, 64)
		if err != nil {
			http.Error(w, "'lon' parameter must be a valid float", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), LatContextKey, lat)
		ctx = context.WithValue(ctx, LonContextKey, lon)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetMethodMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func PostMethodMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ValidateTokenMiddleware(a *api.API) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			token := parts[1]

			tokenValid := a.ValidateToken(token)
			if !tokenValid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UsernameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usernameParam := r.URL.Query().Get("username")

		if usernameParam == "" {
			http.Error(w, "Missing 'username' parameter", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), UsernameContextKey, usernameParam)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
