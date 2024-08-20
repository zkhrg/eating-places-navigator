package elasticsearch

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

type Config struct {
	Address  string
	Username string
	Password string
	Cert     []byte
}

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
	ID       int      `json:"id"`
	Address  string   `json:"address"`
	Location Geopoint `json:"location"`
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
}

type Geopoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type CountResponse struct {
	Count int `json:"count"`
}

func NewClient(cfg *Config) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.Address},
		Username:  cfg.Username,
		Password:  cfg.Password,
		CACert:    cfg.Cert,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %v", err)
	}
	return es, nil
}
