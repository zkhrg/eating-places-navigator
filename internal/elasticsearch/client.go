package elasticsearch

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

var es *elasticsearch.Client

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
