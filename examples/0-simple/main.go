package main

import (
	"fmt"
	"log"

	"github.com/mateothegreat/go-config/config"
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/mateothegreat/go-config/validation"
)

func main() {
	cfg := &AppConfig{}

	err := config.LoadWithPlugins(
		config.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		config.FromEnv(sources.EnvOpts{Prefix: "SIMPLE"}),
	).WithValidationStrategy(validation.StrategyAuto).Build(cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println(cfg)
}
