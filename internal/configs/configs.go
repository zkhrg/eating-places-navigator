package configs

import (
	"os"

	"github.com/zkhrg/go_day03/internal/pkg/elasticsearch"
)

type env string

func (e env) String() string {
	return string(e)
}

const (
	EnvLocal env = "local"
)

// структура Configs обрабатывает все зависимости необходимые для
// обработки конфигураций
type Configs struct {
	Environment env
	AppName     string
	AppVersion  string
}

func (cfg *Configs) Elasticsearch() *elasticsearch.Config {
	return &elasticsearch.Config{
		Address:  os.Getenv("ES_ADDRESS"),
		Username: os.Getenv("ES_USERNAME"),
		Password: os.Getenv("ES_PASSWORD"),
		Cert:     []byte(os.Getenv("ES_CERT_CONTENT")),
	}
}

func loadEnv() env {
	switch env(os.Getenv("ENV")) {
	case EnvLocal:
		return EnvLocal
	default:
		return EnvLocal
	}
}

// New возвращает новый инстанс конфига со всеми необходимыми
// зависимостями инициализованно
func New() (*Configs, error) {
	return &Configs{
		Environment: loadEnv(),
		AppName:     os.Getenv("APP_NAME"),
		AppVersion:  os.Getenv("APP_VERSION"),
	}, nil
}

func (cfg *Configs) PlacesElasticsearchIndex() string {
	return "places"
}
