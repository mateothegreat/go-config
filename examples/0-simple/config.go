package main

type AppConfig struct {
	Name        string    `yaml:"name" validate:"required,minlen=3,maxlen=50"`
	Environment string    `yaml:"environment" validate:"oneof=dev|staging|prod"`
	Version     string    `yaml:"version" validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
	Sub         SubConfig `yaml:"sub" validate:"required"`
}

type SubConfig struct {
	Foo string `yaml:"foo" validate:"required,minlen=3"`
}
