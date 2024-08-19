package httpserver

import "net/http"

// // RegisterRoutes регистрирует все маршруты приложения
// func RegisterRoutes() {
// 	http.HandleFunc("/", StartPageHandler)
// 	http.HandleFunc("/api/", SimpleAPIHandler)
// 	http.HandleFunc("/api/recommend/", APIRecommendHandler)
// 	http.HandleFunc("/api/get_token/", generateTokenHandler)
// 	http.HandleFunc("/protected_endpoint/", protectedEndpoint)
// }

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Use ChainMiddleware to apply middleware to the handler
	handler := ChainMiddleware(
		http.HandlerFunc(itemHandler),
		ValidatePageParamMiddleware,
	)

	mux.Handle("/items", handler)

	return mux
}
