package main

import (
	"github.com/zkhrg/go_day03/internal/configs"
	"github.com/zkhrg/go_day03/internal/pkg/elasticsearch"
)

func main() {
	cfgs, err := configs.New()
	if err != nil {
		return
	}
}
