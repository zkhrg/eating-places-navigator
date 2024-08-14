package main

import "github.com/zkhrg/go_day03/internal/elasticsearch"

func main() {
	elasticsearch.InitClient()
	elasticsearch.CreateIndex("tests_again")
}
