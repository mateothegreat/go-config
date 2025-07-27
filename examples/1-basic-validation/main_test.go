package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/plugins/sources"
)

func TestLoadConfigFluentAPI(t *testing.T) {
	config := &ServerConfig{}

	err := goconfig.LoadWithPlugins(
		goconfig.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
	).Build(config)

	require.NoError(t, err)

	assert.Equal(t, "my-awesome-server", config.Name)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "admin@example.com", config.Email)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, false, config.Debug)
	assert.Equal(t, "v1.2.3", config.Version)
}

func TestLoadConfigLoaderAPI(t *testing.T) {
	config := &ServerConfig{}

	loader := goconfig.NewLoader(config)

	// Add YAML source
	yamlPlugin, err := goconfig.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config.yaml"})
	require.NoError(t, err)
	loader.Use(yamlPlugin)

	// Load and validate
	err = loader.Load(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "my-awesome-server", config.Name)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "admin@example.com", config.Email)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, false, config.Debug)
	assert.Equal(t, "v1.2.3", config.Version)
}

func TestLoadConfigWithDefaults(t *testing.T) {
	config := &ServerConfig{}

	defaults := &ServerConfig{
		Host:     "0.0.0.0",
		Port:     3000,
		LogLevel: "debug",
		Debug:    true,
	}

	err := goconfig.LoadWithPlugins(
		goconfig.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
	).WithDefaults(defaults).Build(config)

	require.NoError(t, err)

	// YAML values should override defaults
	assert.Equal(t, "my-awesome-server", config.Name)
	assert.Equal(t, 8080, config.Port)        // From YAML, not default
	assert.Equal(t, "localhost", config.Host) // From YAML, not default
	assert.Equal(t, "info", config.LogLevel)  // From YAML, not default
	assert.Equal(t, false, config.Debug)      // From YAML, not default
	assert.Equal(t, "v1.2.3", config.Version)
}

func TestLoadConfigWithEnvOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_NAME", "env-server")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_DEBUG", "true")
	defer func() {
		os.Unsetenv("SERVER_NAME")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_DEBUG")
	}()

	config := &ServerConfig{}

	err := goconfig.LoadWithPlugins(
		goconfig.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		goconfig.FromEnv(sources.EnvOpts{Prefix: "SERVER"}),
	).Build(config)

	require.NoError(t, err)

	// Environment variables should override YAML
	assert.Equal(t, "env-server", config.Name) // From env
	assert.Equal(t, 9090, config.Port)         // From env
	assert.Equal(t, true, config.Debug)        // From env
	// Other values from YAML
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "admin@example.com", config.Email)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "v1.2.3", config.Version)
}

func TestValidationWithInvalidConfig(t *testing.T) {
	config := &ServerConfig{
		Name:     "ab",            // Too short (min 3)
		Port:     70000,           // Too high (max 65535)
		Host:     "",              // Required
		Email:    "invalid-email", // Invalid format
		LogLevel: "invalid",       // Not in allowed values
		Debug:    false,
		Version:  "1.2.3", // Invalid format (missing 'v')
	}

	// Use a static source for testing
	loader := goconfig.NewLoader(config)
	err := loader.Load(context.Background())

	assert.Error(t, err)

	// Check if we can get validation errors
	if validationErrs, ok := goconfig.AsValidationErrors(err); ok {
		errors := validationErrs.Errors()
		assert.True(t, len(errors) > 0)
	}
}

func TestInspectionAPI(t *testing.T) {
	config := &ServerConfig{}

	loader := goconfig.NewLoader(config)

	yamlPlugin, err := goconfig.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config.yaml"})
	require.NoError(t, err)
	loader.Use(yamlPlugin)

	err = loader.Load(context.Background())
	require.NoError(t, err)

	// Test sources
	sources := loader.Sources()
	assert.Contains(t, sources, "yaml")

	// Test inspection
	data := loader.Inspect()
	assert.Equal(t, "my-awesome-server", data["name"])
	assert.Equal(t, 8080, data["port"]) // Port number from YAML
	assert.Equal(t, "localhost", data["host"])
}
