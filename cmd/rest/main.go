package main

import "github.com/zkhrg/go_day03/internal/elasticsearch"

func main() {
	mapping := `{
	"settings": {
		"number_of_shards": 1
	},
	"mappings": {
		"properties": {
			"name": {
				"type":  "text"
			},
			"address": {
				"type":  "text"
			},
			"phone": {
				"type":  "text"
			},
			"location": {
				"type": "geo_point"
			}
		}
	}
}`
	elasticsearch.InitClient()
	elasticsearch.CreateIndex("loops", mapping)
}
