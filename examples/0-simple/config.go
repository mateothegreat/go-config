package main

type AppConfig struct {
	Name        string `validate:"required,minlen=3,maxlen=50" yaml:"name"`
	Environment string `validate:"oneof=dev|staging|prod" yaml:"environment"`
	Version     string `validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$" yaml:"version"`
}
