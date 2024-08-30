package main

import (
	"fmt"
	"log"
	"net/http"

	myHttp "github.com/zkhrg/go_day03/cmd/server/http"
	"github.com/zkhrg/go_day03/internal/api"
	"github.com/zkhrg/go_day03/internal/configs"
	"github.com/zkhrg/go_day03/internal/pkg/elasticsearch"
	"github.com/zkhrg/go_day03/internal/places"
)

func main() {
	cfgs, err := configs.New()
	if err != nil {
		// need to handle this case
		return
	}

	es, err := elasticsearch.NewClient(cfgs.Elasticsearch())
	if err != nil {
		fmt.Printf("cannot create new es client\n")
	}
	ess := places.NewElasticsearchStore(es, cfgs.PlacesElasticsearchIndex())
	ess.CreatePlacesIndex()
	ess.IndexingPlaces()
	placesAPI := api.NewStoreAPI(ess)
	mainMux := http.NewServeMux()
	myHttp.AddPlacesRoutes(placesAPI, mainMux)

	log.Println("Starting server on :8888")
	if err := http.ListenAndServe(":8888", mainMux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
