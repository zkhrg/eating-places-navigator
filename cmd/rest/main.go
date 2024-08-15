package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/zkhrg/go_day03/internal/elasticsearch"
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
	// elasticsearch.GetPageData(0, 100, "places")
}
