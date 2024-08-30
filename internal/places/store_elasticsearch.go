package places

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
)

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

type Index struct {
	Index    string        `json:"name"`
	Settings IndexSettings `json:"settings"`
	Mappings IndexMappings `json:"mappings"`
}

type IndexSettings struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

type IndexMappings struct {
	Properties IndexMappingsProperties `json:"properties"`
}

type IndexMappingsProperties struct {
	ID       map[string]string `json:"id"`
	Name     map[string]string `json:"name"`
	Address  map[string]string `json:"address"`
	Phone    map[string]string `json:"phone"`
	Location map[string]string `json:"location"`
}

const batchSize = 500

func (ess *esstore) CreatePlacesIndex() {
	index := Index{
		Index: "places",
		Settings: IndexSettings{
			NumberOfShards:   5,
			NumberOfReplicas: 2,
		},
		Mappings: IndexMappings{
			Properties: IndexMappingsProperties{
				ID:       map[string]string{"type": "keyword"},
				Name:     map[string]string{"type": "text"},
				Address:  map[string]string{"type": "text"},
				Phone:    map[string]string{"type": "keyword"},
				Location: map[string]string{"type": "geo_point"},
			},
		},
	}

	ess.createIndex(index)
}

func (ess *esstore) DeletePlacesIndex() {
	ess.deleteIndex(ess.indexName)
}

func (ess *esstore) IndexingPlaces() {

	headerMap := map[string]string{
		"Name":      "name",
		"Address":   "address",
		"Phone":     "phone",
		"Longitude": "location.lon",
		"Latitude":  "location.lat",
		"ID":        "id",
	}
	file, err := os.Open("./.datasets/data.csv")
	if err != nil {
		log.Fatalf("error opening csv file: %s", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'

	// Чтение заголовков
	headers, err := reader.Read()
	if err != nil {
		log.Fatalf("error reading headers: %s", err)
	}

	var wg sync.WaitGroup

	var batch []map[string]interface{}

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() != "EOF" {
				log.Fatalf("error reading CSV file: %s", err)
			}
			break
		}

		doc := make(map[string]interface{})
		for i, value := range record {
			if i < len(headers) {
				key := headerMap[headers[i]]
				if key == "phone" && value != "" && !strings.HasPrefix(value, "+7") {
					value = "+7" + value
				}
				if key == "id" {
					num, _ := strconv.Atoi(value)
					num += 1
					doc[key] = num
					continue
				}
				if strings.HasPrefix(key, "location.") {
					if doc["location"] == nil {
						doc["location"] = make(map[string]interface{})
					}
					locMap := doc["location"].(map[string]interface{})
					locKey := strings.TrimPrefix(key, "location.")
					valueFloat, _ := strconv.ParseFloat(value, 64)
					locMap[locKey] = valueFloat
				} else {
					doc[key] = value
				}
			}
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			wg.Add(1)
			go func(batch []map[string]interface{}) {
				defer wg.Done()
				ess.sendBatch(batch)
			}(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		wg.Add(1)
		go func(batch []map[string]interface{}) {
			defer wg.Done()
			ess.sendBatch(batch)
		}(batch)
	}

	wg.Wait()

	fmt.Println("data indexing completed.")
}

func (ess *esstore) sendBatch(batch []map[string]interface{}) {
	var buf strings.Builder
	for _, doc := range batch {

		// Формируем метаданные для действия индексации с указанием идентификатора
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : %d } }`, doc["id"]))

		// Преобразуем документ в JSON
		data, _ := json.Marshal(doc)

		// Добавляем метаданные и данные в буфер
		buf.Write(meta)
		buf.WriteByte('\n')
		buf.Write(data)
		buf.WriteByte('\n')
		fmt.Println(doc)
	}

	res, err := ess.esdriver.Bulk(
		strings.NewReader(buf.String()),
		ess.esdriver.Bulk.WithIndex(ess.indexName),
		ess.esdriver.Bulk.WithRefresh("true"),
	)
	if err != nil {
		log.Printf("error indexing batch: %s", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] error indexing batch: %s", res.Status(), res.String())
	} else {
		log.Printf("batch indexed successfully")
	}
}

func (ess *esstore) createIndex(index Index) {
	indexExists, err := ess.esdriver.Indices.Exists([]string{index.Index})
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
	}
	defer indexExists.Body.Close()

	if indexExists.StatusCode == 200 {
		fmt.Println("Index already exists")
		return
	}

	settingsJSON, err := json.Marshal(index.Settings)
	if err != nil {
		log.Fatalf("Error marshalling index settings: %s", err)
	}

	createIndexResponse, err := ess.esdriver.Indices.Create(
		index.Index,
		ess.esdriver.Indices.Create.WithBody(bytes.NewReader(settingsJSON)),
	)
	if err != nil {
		log.Fatalf("Error creating index: %s", err)
	}
	defer createIndexResponse.Body.Close()

	if createIndexResponse.IsError() {
		log.Printf("Error: %s", createIndexResponse.String())
	} else {
		fmt.Println("Index created successfully")
	}
}

func (ess *esstore) deleteIndex(indexName string) error {
	res, err := ess.esdriver.Indices.Delete([]string{indexName})
	if err != nil {
		return fmt.Errorf("error deleting index: %s", err)
	}
	defer res.Body.Close()
	return nil
}
