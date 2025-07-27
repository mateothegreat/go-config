package main

import (
	"fmt"
	"log"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/plugins/sources"
)

func main() {
	cfg := &AppConfig{}

	err := goconfig.LoadWithPlugins(
		goconfig.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		goconfig.FromEnv(sources.EnvOpts{Prefix: "APP"}),
	).WithValidationStrategy(goconfig.StrategyAuto).Build(cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println(cfg)
}
