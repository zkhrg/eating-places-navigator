package http

import (
	"net/http"

	"github.com/zkhrg/go_day03/internal/api"
)

func AddPlacesRoutes(a *api.API, mux *http.ServeMux) {
	// Создаем цепочку миддлварей и передаем API через замыкание
	HTMLPaginatedChain := ChainMiddleware(
		HTMLPageHandler(a),      // Передаем API в хендлер html-ки
		GetMethodMiddleware,     // первое что мы делаем это миддлеварь на гет запрос
		PaginationMiddleware(a), // затем проверяем что у нас предоставлен page
	)

	JSONPaginatedChain := ChainMiddleware(
		JSONPageHandler(a),
		GetMethodMiddleware,
		PaginationMiddleware(a),
	)

	JSONRecommendChain := ChainMiddleware(
		NearestPlacesHandler(a),
		GetMethodMiddleware,
		LatLonMiddleware,
	)

	mux.Handle("/api/recommend", JSONRecommendChain)
	mux.Handle("/api/places", JSONPaginatedChain)
	mux.Handle("/", HTMLPaginatedChain)
}
