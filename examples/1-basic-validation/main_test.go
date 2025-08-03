package main

import (
	"context"
	"os"
	"testing"

	"github.com/mateothegreat/go-config/config"
	"github.com/mateothegreat/go-config/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mateothegreat/go-config/plugins"
	"github.com/mateothegreat/go-config/plugins/sources"
)

func TestLoadConfigFluentAPI(t *testing.T) {
	cfg := &ServerConfig{}

	err := config.LoadWithPlugins(
		config.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
	).Build(cfg)

	require.NoError(t, err)

	assert.Equal(t, "my-awesome-server", cfg.Name)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "admin@example.com", cfg.Email)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, "v1.2.3", cfg.Version)
}

func TestLoadConfigLoaderAPI(t *testing.T) {
	cfg := &ServerConfig{}

	loader := config.NewConfigLoader(cfg)

	// Add YAML source.
	yamlPlugin, err := plugins.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config.yaml"})
	require.NoError(t, err)
	loader.Use(yamlPlugin)

	// Load and validate.
	err = loader.Load(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "my-awesome-server", cfg.Name)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "admin@example.com", cfg.Email)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, "v1.2.3", cfg.Version)
}

func TestLoadConfigWithDefaults(t *testing.T) {
	cfg := &ServerConfig{}

	defaults := &ServerConfig{
		Host:     "0.0.0.0",
		Port:     3000,
		LogLevel: "debug",
		Debug:    true,
	}

	err := config.LoadWithPlugins(
		config.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
	).WithDefaults(defaults).Build(cfg)

	require.NoError(t, err)

	// YAML values should override defaults.
	assert.Equal(t, "my-awesome-server", cfg.Name)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, "v1.2.3", cfg.Version)
}

func TestLoadConfigWithEnvOverride(t *testing.T) {
	// Set environment variables.
	os.Setenv("SERVER_NAME", "env-server")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_DEBUG", "true")
	defer func() {
		os.Unsetenv("SERVER_NAME")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_DEBUG")
	}()

	cfg := &ServerConfig{}

	err := config.LoadWithPlugins(
		config.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		config.FromEnv(sources.EnvOpts{Prefix: "SERVER"}),
	).Build(cfg)

	require.NoError(t, err)

	// Environment variables should override YAML.
	assert.Equal(t, "env-server", cfg.Name)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, true, cfg.Debug)
	// Other values from YAML.
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "admin@example.com", cfg.Email)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "v1.2.3", cfg.Version)
}

func TestValidationWithInvalidConfig(t *testing.T) {
	cfg := &ServerConfig{
		Name:     "ab",            // Too short (min 3)
		Port:     70000,           // Too high (max 65535)
		Host:     "",              // Required
		Email:    "invalid-email", // Invalid format
		LogLevel: "invalid",       // Not in allowed values
		Debug:    false,
		Version:  "1.2.3", // Invalid format (missing 'v')
	}

	// Use a static source for testing
	loader := config.NewConfigLoader(cfg)
	err := loader.Load(context.Background())

	assert.Error(t, err)

	// Check if we can get validation errors.
	if validationErrs, ok := errors.AsValidationErrors(err); ok {
		errors := validationErrs.Errors()
		assert.True(t, len(errors) > 0)
	}
}

func TestInspectionAPI(t *testing.T) {
	cfg := &ServerConfig{}

	loader := config.NewConfigLoader(cfg)

	yamlPlugin, err := plugins.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config.yaml"})
	require.NoError(t, err)
	loader.Use(yamlPlugin)

	err = loader.Load(context.Background())
	require.NoError(t, err)

	// Test sources.
	sources := loader.Sources()
	assert.Contains(t, sources, "yaml")

	// Test inspection.
	data := loader.Inspect()
	assert.Equal(t, "my-awesome-server", data["name"])
	assert.Equal(t, 8080, data["port"])
	assert.Equal(t, "localhost", data["host"])
}
