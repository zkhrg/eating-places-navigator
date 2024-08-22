package elasticsearch

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

const batchSize = 500

type Config struct {
	Address  string
	Username string
	Password string
	Cert     []byte
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
