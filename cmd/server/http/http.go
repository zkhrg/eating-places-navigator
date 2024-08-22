package http

import (
	"net/http"

	"github.com/zkhrg/go_day03/internal/api"
)

func NewRouter(a *api.API) *http.ServeMux {
	mux := http.NewServeMux()

	// Создаем цепочку миддлварей и передаем API через замыкание
	HTMLpaginatedChain := ChainMiddleware(
		HTMLPageHandler(a), // Передаем API в хендлер
		PaginationMiddleware(a),
	)

	JSONpaginatedChain := ChainMiddleware(
		JSONPageHandler(a),
		PaginationMiddleware(a),
	)

	mux.Handle("GET /api/places", JSONpaginatedChain)
	mux.Handle("GET /", HTMLpaginatedChain)

	return mux
}
