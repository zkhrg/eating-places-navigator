package index

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
