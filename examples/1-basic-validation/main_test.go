package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfigFromFile("config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	assert.NoError(t, err)

	assert.Equal(t, config.Name, "my-awesome-server")
	assert.Equal(t, config.Port, 8080)
	assert.Equal(t, config.Host, "localhost")
	assert.Equal(t, config.Email, "admin@example.com")
	assert.Equal(t, config.LogLevel, "info")
	assert.Equal(t, config.Debug, false)
	assert.Equal(t, config.Version, "v1.2.3")
}
