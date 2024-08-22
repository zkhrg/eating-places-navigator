package places

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

const batchSize = 500

type SearchResponse struct {
	Hits struct {
		Hits []PlacesHit `json:"hits"`
	} `json:"hits"`
}

type PlacesHit struct {
	Source Place         `json:"_source"`
	Sort   []interface{} `json:"sort"`
}

type CountResponse struct {
	Count int `json:"count"`
}

type esstore struct {
	esdriver  *elasticsearch.Client
	indexName string
}

func (ess *esstore) GetPlacesByPageParams(ctx context.Context, pageNumber int, pageSize int) ([]Place, error) {
	searchAfter := 0
	var r SearchResponse
	chunkSize := pageSize
	recordsCount := ess.GetTotalRecords()
	if recordsCount == 0 {
		return nil, nil
	}
	chunkSize = correctChunkSize(chunkSize)
	chunkPagesNumber := pageNumber/(chunkSize/pageSize) + 1

	for i := 0; i < chunkPagesNumber; i++ {
		searchBody := map[string]interface{}{
			"search_after": []interface{}{searchAfter},
			"size":         chunkSize,
			"sort": []map[string]interface{}{
				{"id": "asc"},
			},
			"track_total_hits": true,
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
			log.Fatalf("Error encoding query: %s", err)
		}

		res, err := ess.esdriver.Search(
			ess.esdriver.Search.WithIndex(ess.indexName),
			ess.esdriver.Search.WithBody(&buf),
		)
		if err != nil {
			log.Fatalf("Error getting the response: %s", err)
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
			return nil, nil
		}
		searchAfter = int(r.Hits.Hits[len(r.Hits.Hits)-1].Sort[0].(float64))
	}
	start, end := calcStartEndForPage(pageSize, pageNumber, chunkSize, recordsCount)
	return placesHitsToPlaces(r.Hits.Hits[start:end]), nil
}

func (ess *esstore) GetTotalRecords() int {
	var rc CountResponse
	res_count, _ := ess.esdriver.Count(
		ess.esdriver.Count.WithIndex(ess.indexName),
	)
	if err := json.NewDecoder(res_count.Body).Decode(&rc); err != nil {
		log.Fatalf("error parsing the response body count: %s", err)
		return 0
	}
	return rc.Count
}

func correctChunkSize(chunkSize int) int {
	mult := 1
	if chunkSize < 100 {
		mult = 100
	} else if chunkSize < 1000 {
		mult = 10
	}
	chunkSize *= mult
	return chunkSize
}

func calcStartEndForPage(pageSize, pageNumber, chunkSize, recordsCount int) (int, int) {
	start := (pageSize * (pageNumber - 1)) % chunkSize
	end := start + pageSize

	if end >= recordsCount%chunkSize {
		end = (recordsCount % chunkSize) - 1
	}
	return start, end
}

func placesHitsToPlaces(ph []PlacesHit) []Place {
	res := make([]Place, len(ph))
	for i, v := range ph {
		res[i].Name = v.Source.Name
		res[i].Phone = v.Source.Phone
		res[i].ID = v.Source.ID
		res[i].Address = v.Source.Address
		res[i].Location.Lat = v.Source.Location.Lat
		res[i].Location.Lon = v.Source.Location.Lon
	}
	return res
}

func (ess *esstore) GetNearestPlaces(lat, lon float64) ([]Place, error) {
	searchBody := map[string]interface{}{
		"size": 3,
		"sort": []map[string]interface{}{
			{"_geo_distance": map[string]interface{}{
				"location": map[string]interface{}{
					"lat": lat,
					"lon": lon,
				},
				"order":           "asc",
				"unit":            "km",
				"mode":            "min",
				"distance_type":   "arc",
				"ignore_unmapped": true,
			}},
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		log.Fatalf("Error encoding query: %s", err)
		return nil, err
	}

	res, err := ess.esdriver.Search(
		ess.esdriver.Search.WithIndex(ess.indexName),
		ess.esdriver.Search.WithBody(&buf),
	)
	if err != nil {
		log.Fatalf("Error getting the response: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
		return nil, err
	}
	return placesHitsToPlaces(r.Hits.Hits), nil
}

func NewElasticsearchStore(esdriver *elasticsearch.Client, indexName string) *esstore {
	return &esstore{
		indexName: indexName,
		esdriver:  esdriver,
	}
}

// package elasticsearch

// import (
// 	"bytes"
// 	"encoding/csv"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"sync"

// 	"github.com/elastic/go-elasticsearch/v8"
// 	"github.com/joho/godotenv"
// )

// var es *elasticsearch.Client

// func InitClient() {
// 	var err error
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatalf("Error loading .env file: %s", err)
// 	}

// 	esUsername := os.Getenv("ES_USERNAME")
// 	esPassword := os.Getenv("ES_PASSWORD")
// 	esAddress := os.Getenv("ES_ADDRESS")
// 	esCertPath := os.Getenv("ES_CERT_PATH")

// 	cert, err := os.ReadFile(esCertPath)
// 	if err != nil {
// 		fmt.Printf("error reading https_ca.crt: %s", err)
// 	}

// 	cfg := elasticsearch.Config{
// 		Addresses: []string{
// 			esAddress,
// 		},
// 		Username: esUsername,
// 		Password: esPassword,
// 		CACert:   cert,
// 	}

// 	es, err = elasticsearch.NewClient(cfg)
// 	if err != nil {
// 		fmt.Printf("error creating the client: %s", err)
// 	}
// }

// func GetPageData(pageNumber int, pageSize int, indexName string) []PlacesHit {
// 	searchAfter := 0
// 	var r SearchResponse
// 	chunkSize := pageSize
// 	recordsCount := CountIndexRecords(indexName)
// 	if recordsCount == 0 {
// 		return nil
// 	}
// 	mult := 1
// 	if pageSize < 100 {
// 		mult = 100
// 	} else if pageSize < 1000 {
// 		mult = 10
// 	}
// 	chunkSize *= mult

// 	chunkPagesNumber := pageNumber/(chunkSize/pageSize) + 1

// 	for i := 0; i < chunkPagesNumber; i++ {
// 		searchBody := map[string]interface{}{
// 			"search_after": []interface{}{searchAfter},
// 			"size":         chunkSize,
// 			"sort": []map[string]interface{}{
// 				{"id": "asc"},
// 			},
// 			"track_total_hits": true,
// 			"query": map[string]interface{}{
// 				"match_all": map[string]interface{}{},
// 			},
// 		}

// 		var buf bytes.Buffer
// 		if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
// 			log.Fatalf("Error encoding query: %s", err)
// 		}

// 		res, err := es.Search(
// 			es.Search.WithIndex(indexName),
// 			es.Search.WithBody(&buf),
// 		)
// 		if err != nil {
// 			log.Fatalf("Error getting the response: %s", err)
// 		}
// 		defer res.Body.Close()

// 		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
// 			log.Fatalf("Error parsing the response body: %s", err)
// 			return nil
// 		}
// 		searchAfter = int(r.Hits.Hits[len(r.Hits.Hits)-1].Sort[0].(float64))
// 	}

// 	start := (pageSize * (pageNumber - 1)) % chunkSize
// 	end := start + pageSize

// 	if end >= recordsCount%chunkSize {
// 		end = (recordsCount % chunkSize) - 1
// 	}

// 	pages := recordsCount / pageSize
// 	if recordsCount%pageSize != 0 {
// 		pages += 1
// 	}
// 	return r.Hits.Hits[start:end]
// }

// func CountIndexRecords(indexName string) int {
// 	var rc CountResponse
// 	res_count, _ := es.Count(
// 		es.Count.WithIndex(indexName),
// 	)
// 	if err := json.NewDecoder(res_count.Body).Decode(&rc); err != nil {
// 		log.Fatalf("error parsing the response body count: %s", err)
// 		return 0
// 	}
// 	return rc.Count
// }

// func GetNearestPlaces(lat float64, lon float64, indexName string) []PlacesHit {
// 	fmt.Println(lat, lon)
// 	searchBody := map[string]interface{}{
// 		"size": 3,
// 		"sort": []map[string]interface{}{
// 			{"_geo_distance": map[string]interface{}{
// 				"location": map[string]interface{}{
// 					"lat": lat,
// 					"lon": lon,
// 				},
// 				"order":           "asc",
// 				"unit":            "km",
// 				"mode":            "min",
// 				"distance_type":   "arc",
// 				"ignore_unmapped": true,
// 			}},
// 		},
// 	}
// 	var buf bytes.Buffer
// 	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
// 		log.Fatalf("Error encoding query: %s", err)
// 	}

// 	res, err := es.Search(
// 		es.Search.WithIndex(indexName),
// 		es.Search.WithBody(&buf),
// 	)
// 	if err != nil {
// 		log.Fatalf("Error getting the response: %s", err)
// 	}
// 	defer res.Body.Close()

// 	var r SearchResponse
// 	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
// 		log.Fatalf("Error parsing the response body: %s", err)
// 		return nil
// 	}
// 	return r.Hits.Hits
// }
