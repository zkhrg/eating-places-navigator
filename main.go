package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zkhrg/go_day03/internal/api"
	"github.com/zkhrg/go_day03/internal/configs"
	"github.com/zkhrg/go_day03/internal/pkg/elasticsearch"
	"github.com/zkhrg/go_day03/internal/places"
)

func main() {
	ctx := context.Background()
	cfgs, err := configs.New()
	if err != nil {
		// need to handle this case
		return
	}
	es, err := elasticsearch.NewClient(cfgs.Elasticsearch())
	if err != nil {
		fmt.Printf("cannot create new es client\n")
	}
	placesStore := places.NewElasticsearchStore(es, cfgs.PlacesElasticsearchIndex())
	placesAPI := api.New(placesStore)
	page, _ := placesAPI.GetPage(ctx, 1, 5)
	prettyPage, _ := json.MarshalIndent(page, "", "    ")
	fmt.Println(string(prettyPage))
}
