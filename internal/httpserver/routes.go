package httpserver

import "net/http"

// RegisterRoutes регистрирует все маршруты приложения
func RegisterRoutes() {
	http.HandleFunc("/", StartPageHandler)
}
