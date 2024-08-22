package main

import (
	"bytes"
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

type Index struct {
	Name     string
	Settings IndexSettings
}

type IndexSettings struct {
	NumberOfShards   int                   `json:"number_of_shards"`
	NumberOfReplicas int                   `json:"number_of_replicas"`
	Mappings         IndexSettingsMappings `json:"mappings"`
}

type IndexSettingsMappings struct {
	Properties IndexSettingsMappingsProperties `json:"properties"`
}

type IndexSettingsMappingsProperties struct {
	ID       map[string]string `json:"id"`
	Name     map[string]string `json:"name"`
	Address  map[string]string `json:"address"`
	Phone    map[string]string `json:"phone"`
	Location map[string]string `json:"location"`
}

const batchSize = 500

type esstore struct {
	esdriver  *elasticsearch.Client
	indexName string
}

func NewElasticsearchStore(esdriver *elasticsearch.Client, indexName string) *esstore {
	return &esstore{
		indexName: indexName,
		esdriver:  esdriver,
	}
}

func (ess *esstore) CreatePlacesIndex() {
	index := Index{
		Name: "places",
		Settings: IndexSettings{
			NumberOfShards:   5,
			NumberOfReplicas: 2,
			Mappings: IndexSettingsMappings{
				Properties: IndexSettingsMappingsProperties{
					ID:       map[string]string{"type": "keyword"},
					Name:     map[string]string{"type": "text"},
					Address:  map[string]string{"type": "text"},
					Phone:    map[string]string{"type": "keyword"},
					Location: map[string]string{"type": "geo_point"},
				},
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
	file, err := os.Open("./datasets/data.csv")
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
	indexExists, err := ess.esdriver.Indices.Exists([]string{index.Name})
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
		index.Name,
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
