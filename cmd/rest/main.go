package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/zkhrg/go_day03/internal/elasticsearch"
	"github.com/zkhrg/go_day03/internal/httpserver"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}
	modulePath := os.Getenv("MODULE_PATH")
	mapping, err := os.ReadFile(modulePath + "internal/elasticsearch/schemas/places_mapping.json")
	if err != nil {
		log.Fatalf("Error reading mapping file: %s", err)
	}
	elasticsearch.InitClient()
	elasticsearch.CreateIndex("places", string(mapping))
	_ = mapping
	elasticsearch.Indexing("places", modulePath+"datasets/data.csv")
	// elasticsearch.GetPageData(13444, 1, "places")

	httpserver.RegisterRoutes()

	http.ListenAndServe("localhost:8888", nil)

}
