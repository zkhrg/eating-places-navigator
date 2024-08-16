package elasticsearch

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
	"github.com/joho/godotenv"
)

var es *elasticsearch.Client

const batchSize = 500

type SearchResponse struct {
	Hits struct {
		Hits []PlacesHit `json:"hits"`
	} `json:"hits"`
}

type PlacesHit struct {
	Source IndexEntry    `json:"_source"`
	Sort   []interface{} `json:"sort"`
}

type IndexEntry struct {
	ID       string   `json:"id"`
	Address  string   `json:"address"`
	Location Geopoint `json:"location"`
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
}

type Geopoint struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type CountResponse struct {
	Count int `json:"count"`
}

func InitClient() {
	var err error
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	esUsername := os.Getenv("ES_USERNAME")
	esPassword := os.Getenv("ES_PASSWORD")
	esAddress := os.Getenv("ES_ADDRESS")
	esCertPath := os.Getenv("ES_CERT_PATH")

	cert, err := os.ReadFile(esCertPath)
	if err != nil {
		fmt.Printf("error reading https_ca.crt: %s", err)
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			esAddress,
		},
		Username: esUsername,
		Password: esPassword,
		CACert:   cert,
	}

	es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Printf("error creating the client: %s", err)
	}
}

func CreateIndex(indexName string, mapping string) error {
	exists, err := es.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("error check index existence: %s", err)
	}
	if exists.StatusCode == 200 {
		return nil
	}

	res, err := es.Indices.Create(
		indexName,
		es.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return fmt.Errorf("error creating index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from elasticsearch: %s", res.String())
	}

	return nil
}

func DeleteIndex(indexName string) error {
	res, err := es.Indices.Delete([]string{indexName})
	if err != nil {
		return fmt.Errorf("error deleting index: %s", err)
	}
	defer res.Body.Close()
	return nil
}

func Indexing(indexName string, CSVFileName string) {
	headerMap := map[string]string{
		"Name":      "name",
		"Address":   "address",
		"Phone":     "phone",
		"Longitude": "location.lon",
		"Latitude":  "location.lat",
		"ID":        "id",
	}
	file, err := os.Open(CSVFileName)
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
					value = fmt.Sprintf("%d", num)
				}
				if strings.HasPrefix(key, "location.") {
					if doc["location"] == nil {
						doc["location"] = make(map[string]interface{})
					}
					locMap := doc["location"].(map[string]interface{})
					locKey := strings.TrimPrefix(key, "location.")
					locMap[locKey] = value
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
				sendBatch(es, indexName, batch)
			}(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		wg.Add(1)
		go func(batch []map[string]interface{}) {
			defer wg.Done()
			sendBatch(es, indexName, batch)
		}(batch)
	}

	wg.Wait()

	fmt.Println("data indexing completed.")
}

func sendBatch(es *elasticsearch.Client, indexName string, batch []map[string]interface{}) {
	var buf strings.Builder
	for _, doc := range batch {

		// Формируем метаданные для действия индексации с указанием идентификатора
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }`, doc["id"]))

		// Преобразуем документ в JSON
		data, _ := json.Marshal(doc)

		// Добавляем метаданные и данные в буфер
		buf.Write(meta)
		buf.WriteByte('\n')
		buf.Write(data)
		buf.WriteByte('\n')
		fmt.Println(doc)
	}

	res, err := es.Bulk(
		strings.NewReader(buf.String()),
		es.Bulk.WithIndex(indexName),
		es.Bulk.WithRefresh("true"),
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

func GetPageData(pageNumber int, pageSize int, indexName string) ([]PlacesHit, int) {
	searchAfter := 0
	var r SearchResponse
	var rc CountResponse
	chunkSize := pageSize
	mult := 1
	if pageSize < 100 {
		mult = 100
	} else if pageSize < 1000 {
		mult = 10
	}
	chunkSize *= mult

	chunkPagesNumber := pageNumber/(chunkSize/pageSize) + 1
	res_count, _ := es.Count(
		es.Count.WithIndex(indexName),
	)
	if err := json.NewDecoder(res_count.Body).Decode(&rc); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
		return nil, 0
	}
	i := 0
	for ; i < chunkPagesNumber; i++ {
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

		res, err := es.Search(
			es.Search.WithIndex(indexName),
			es.Search.WithBody(&buf),
		)
		if err != nil {
			log.Fatalf("Error getting the response: %s", err)
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
			return nil, 0
		}
		searchAfter = int(r.Hits.Hits[len(r.Hits.Hits)-1].Sort[0].(float64))
	}

	// Обрабатываем записи
	start := (pageSize * (pageNumber - 1)) % chunkSize
	end := start + pageSize

	if end >= rc.Count%chunkSize {
		end = (rc.Count % chunkSize) - 1
	}
	// вместо вывода можно просто вовзращать из функции ссылкой
	// for _, hit := range r.Hits.Hits[start:end] {
	// 	fmt.Printf("Record: %v %v\n", hit.Source, hit.Sort[0])
	// }
	pages := rc.Count / pageSize
	if rc.Count%pageSize != 0 {
		pages += 1
	}
	return r.Hits.Hits[start:end], rc.Count
}
