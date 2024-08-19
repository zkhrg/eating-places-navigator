package http

import (
	"log"
	"net/http"
	"your_project/internal/configs"
	"your_project/pkg/factory"
)

func StartHTTPServer() {
	cfg := configs.LoadConfig()

	store, err := factory.NewStore(cfg)
	if err != nil {
		log.Fatalf("Error initializing store: %v", err)
	}

	mux := http.NewServeMux()
	RegisterHandlers(mux, store)

	log.Println("Starting HTTP server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
