package main

import (
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/zkhrg/go_day03/docs" // Путь к сгенерированным файлам документации

	myHttp "github.com/zkhrg/go_day03/cmd/server/http"
	"github.com/zkhrg/go_day03/internal/api"
	"github.com/zkhrg/go_day03/internal/configs"
	"github.com/zkhrg/go_day03/internal/pkg/elasticsearch"
	"github.com/zkhrg/go_day03/internal/places"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Type `Bearer <token>` to auth.

func main() {
	cfgs, err := configs.New()
	if err != nil {
		return
	}

	es, err := elasticsearch.NewClient(cfgs.Elasticsearch())
	for err != nil {
		log.Printf("cannot create new es client retry after 5 sec\n")
		time.Sleep(5 * time.Second)
		es, err = elasticsearch.NewClient(cfgs.Elasticsearch())
	}
	ess := places.NewElasticsearchStore(es, cfgs.PlacesElasticsearchIndex())
	ess.CreatePlacesIndex()
	ess.IndexingPlaces()
	placesAPI := api.NewStoreAPI(ess)
	mainMux := http.NewServeMux()
	myHttp.AddPlacesRoutes(placesAPI, mainMux)

	// // handle not existsing pages after add all routes
	// mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	http.NotFound(w, r)
	// })

	mainMux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Обработчик для Swagger JSON
	mainMux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	log.Println("Starting server on :8888")
	if err := http.ListenAndServe(":8888", mainMux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
