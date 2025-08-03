package config

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/mateothegreat/go-config/errors"
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/mateothegreat/go-validation"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
)

// Embedded struct definitions
type BaseConfig struct {
	Host string `yaml:"host" validate:"required" default:"localhost"`
	Port int    `yaml:"port" validate:"required,min=1,max=65535" default:"bad"`
}

type AuthConfig struct {
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type AppConfig struct {
	Name        string     `yaml:"name" validate:"required,minlen=3,maxlen=50"`
	Environment string     `yaml:"environment" validate:"oneof=dev|staging|prod"`
	Version     string     `yaml:"version" validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
	Sub         *SubConfig `yaml:"sub" validate:"required"`
	// Test embedded structs - this should cause the issue
	Redis RedisConfig `yaml:"redis" validate:"required"`
}

type SubConfig struct {
	Foo string `yaml:"foo" validate:"required,minlen=3"`
}

// RedisConfig with embedded structs
type RedisConfig struct {
	BaseConfig        // Embedded - fields should be flattened
	AuthConfig        // Embedded - fields should be flattened
	Database   int    `yaml:"database" validate:"min=0,max=15"`
	Timeout    string `yaml:"timeout" validate:"required"`
}

func TestConfig(t *testing.T) {
	cfg := &AppConfig{}

	err := LoadWithPlugins(
		FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		FromEnv(sources.EnvOpts{Prefix: "SIMPLE"}),
	).Build(cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	assert.NoError(t, err)
	assert.Equal(t, cfg.Name, "go-validate-demo")
	assert.Equal(t, cfg.Environment, "dev")
	assert.Equal(t, cfg.Version, "v1.0.0")
	assert.NotNil(t, cfg.Sub, "pointer struct should not be nil")
	assert.Equal(t, cfg.Sub.Foo, "test-value")
	assert.Equal(t, cfg.Redis.Host, "localhost")
	assert.Equal(t, cfg.Redis.Port, 6379)
	assert.Equal(t, cfg.Redis.Username, "default")
	litter.Dump(cfg)
}

func TestPointerStructValidation(t *testing.T) {
	// Test case where pointer struct is missing (should fail validation)
	cfg := &AppConfig{
		Name:        "test-app",
		Environment: "dev",
		Version:     "v1.0.0",
		Sub:         nil, // This should fail validation due to "required" tag
		Redis: RedisConfig{
			BaseConfig: BaseConfig{Host: "localhost", Port: 6379},
			AuthConfig: AuthConfig{Username: "user", Password: "pass"},
			Timeout:    "5s",
		},
	}

	err := validation.New().Struct(cfg)
	assert.Error(t, err, "validation should fail when required pointer struct is nil")
	assert.Contains(t, err.Error(), "Sub", "error should mention the Sub field")

	// Test case where pointer struct is present and valid
	cfg.Sub = &SubConfig{Foo: "valid-value"}
	err = validation.New().Struct(cfg)
	assert.NoError(t, err, "validation should pass when pointer struct is properly set")
}

func TestDefaultValues(t *testing.T) {
	// Test struct with only partial configuration data (should use defaults)
	cfg := &AppConfig{}

	// Create a data map with some missing fields that have defaults
	data := map[string]any{
		"name":        "test-app",
		"environment": "dev",
		"version":     "v1.0.0",
		"sub": map[string]any{
			"foo": "test-value",
		},
		"redis": map[string]any{
			// host and port are missing - should use defaults from struct tags
			"username": "testuser",
			"password": "testpass",
			"timeout":  "10s",
		},
	}

	hydrator := NewHydrator(HydrationAuto)
	err := hydrator.Hydrate(data, cfg)
	assert.NoError(t, err, "hydration should succeed with defaults")

	// Verify default values were applied
	assert.Equal(t, "localhost", cfg.Redis.Host, "should use default value for Host")
	assert.Equal(t, 6379, cfg.Redis.Port, "should use default value for Port")

	// Verify explicit values were used
	assert.Equal(t, "testuser", cfg.Redis.Username, "should use explicit value for Username")
	assert.Equal(t, "testpass", cfg.Redis.Password, "should use explicit value for Password")
	assert.Equal(t, "10s", cfg.Redis.Timeout, "should use explicit value for Timeout")

	// Test validation with defaults - should pass
	validator := validation.New()
	err = validator.Struct(cfg)
	assert.NoError(t, err, "validation should pass with default values")
}

func TestDefaultValuesValidation(t *testing.T) {
	// Test struct with invalid default value (should fail validation)
	type TestConfig struct {
		Port int `yaml:"port" validate:"min=1,max=65535" default:"70000"` // Invalid default (too high)
	}

	cfg := &TestConfig{}
	data := map[string]any{} // Empty data, should use default

	hydrator := NewHydrator(HydrationAuto)
	err := hydrator.Hydrate(data, cfg)
	assert.NoError(t, err, "hydration should succeed even with invalid default")
	assert.Equal(t, 70000, cfg.Port, "should set the default value")

	// Validation should fail because default value violates constraints
	validator := validation.New()
	err = validator.Struct(cfg)
	assert.Error(t, err, "validation should fail when default value violates constraints")
	assert.Contains(t, err.Error(), "Port", "error should mention the Port field")
}

func TestDefaultValuesWithLoader(t *testing.T) {
	// Test that defaults work when using the full loader pipeline
	type SimpleConfig struct {
		Host string `yaml:"host" validate:"required" default:"localhost"`
		Port int    `yaml:"port" validate:"required,min=1,max=65535" default:"8080"`
		Name string `yaml:"name" validate:"required"`
	}

	cfg := &SimpleConfig{}

	// Create an in-memory source with partial data (missing host and port)
	data := map[string]any{
		"name": "test-service",
	}

	// Use a mock source that returns our partial data
	mockSource := &mockConfigSource{data: data}

	err := NewConfigLoader(cfg).Use(mockSource).Load(context.Background())
	assert.NoError(t, err, "loading should succeed with defaults")

	// Verify defaults were applied
	assert.Equal(t, "localhost", cfg.Host, "should use default value for Host")
	assert.Equal(t, 8080, cfg.Port, "should use default value for Port")
	assert.Equal(t, "test-service", cfg.Name, "should use explicit value for Name")
}

// Mock source for testing
type mockConfigSource struct {
	data map[string]any
}

func (m *mockConfigSource) Name() string {
	return "mock"
}

func (m *mockConfigSource) Load(_ context.Context) (map[string]any, error) {
	return m.data, nil
}

func TestDefaultValuesWithYAMLLoader(t *testing.T) {
	// Test that defaults work with YAML file that has missing values
	cfg := &AppConfig{}

	err := LoadWithPlugins(
		FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
	).Build(cfg)

	assert.NoError(t, err, "loading should succeed with defaults from YAML")

	// Verify defaults were applied for missing values in YAML (host/port are commented out)
	assert.Equal(t, "localhost", cfg.Redis.Host, "should use default value for Host from struct tag")
	assert.Equal(t, 6379, cfg.Redis.Port, "should use default value for Port from struct tag")

	// Verify explicit values from YAML were used
	assert.Equal(t, "default", cfg.Redis.Username, "should use YAML value for Username")
	assert.Equal(t, "secret", cfg.Redis.Password, "should use YAML value for Password")
	assert.Equal(t, "5s", cfg.Redis.Timeout, "should use YAML value for Timeout")

	// Verify other fields
	assert.Equal(t, "go-validate-demo", cfg.Name, "should use YAML value for Name")
	assert.Equal(t, "dev", cfg.Environment, "should use YAML value for Environment")
	assert.Equal(t, "v1.0.0", cfg.Version, "should use YAML value for Version")
	assert.NotNil(t, cfg.Sub, "Sub should be loaded from YAML")
	assert.Equal(t, "test-value", cfg.Sub.Foo, "should use YAML value for Sub.Foo")
}

func TestErrorHandling(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		// Test configuration with validation errors
		type TestConfig struct {
			Port int    `yaml:"port" validate:"min=1,max=65535"`
			Name string `yaml:"name" validate:"required"`
		}

		cfg := &TestConfig{}
		mockSource := &mockConfigSource{
			data: map[string]any{
				"port": 70000, // Invalid - too high
				// name is missing - required field
			},
		}

		err := NewConfigLoader(cfg).Use(mockSource).Load(context.Background())
		assert.Error(t, err, "should return error for validation failures")

		// Check that it's our new error type
		if configErr, ok := err.(*errors.ConfigLoadError); ok {
			assert.True(t, configErr.HasErrors(), "should have errors")

			// Check error message is readable
			errMsg := configErr.Error()
			assert.Contains(t, errMsg, "configuration loading failed", "should have clear error prefix")
			assert.Contains(t, errMsg, "validation failed", "should mention validation failure")

			// Check suggestions are helpful
			suggestions := configErr.GetSuggestions()
			assert.Greater(t, len(suggestions), 0, "should provide suggestions")
		} else {
			t.Fatalf("Expected ConfigLoadError, got %T", err)
		}
	})

	t.Run("SourceErrors", func(t *testing.T) {
		// Test configuration with source errors
		cfg := &AppConfig{}

		err := LoadWithPlugins(
			FromYAML(sources.YAMLOpts{Path: "nonexistent.yaml"}),
		).Build(cfg)

		assert.Error(t, err, "should return error for missing file")

		// Check error message
		errMsg := err.Error()
		assert.Contains(t, errMsg, "configuration loading failed", "should have clear error prefix")
		assert.Contains(t, errMsg, "source errors", "should mention source errors")
	})

	t.Run("NoErrors", func(t *testing.T) {
		// Test that no errors returns nil
		cfg := &AppConfig{}
		mockSource := &mockConfigSource{
			data: map[string]any{
				"name":        "test-app",
				"environment": "dev",
				"version":     "v1.0.0",
				"sub": map[string]any{
					"foo": "test-value",
				},
				"redis": map[string]any{
					"username": "user",
					"password": "pass",
					"timeout":  "5s",
				},
			},
		}

		err := NewConfigLoader(cfg).Use(mockSource).Load(context.Background())
		assert.NoError(t, err, "should not return error when configuration is valid")

		// Verify loader.Errors() also returns nil when there are no errors
		loader := NewConfigLoader(cfg).Use(mockSource)
		_ = loader.Load(context.Background())
		loaderErr := loader.Errors()
		assert.NoError(t, loaderErr, "Errors() should return nil when there are no errors")
	})

	t.Run("HydrationErrors", func(t *testing.T) {
		// Test hydration errors (e.g., type conversion failures)
		type TestConfig struct {
			Port int `yaml:"port"`
		}

		cfg := &TestConfig{}
		mockSource := &mockConfigSource{
			data: map[string]any{
				"port": "invalid-number", // String that can't be converted to int
			},
		}

		err := NewConfigLoader(cfg).Use(mockSource).Load(context.Background())
		assert.Error(t, err, "should return error for hydration failures")

		if configErr, ok := err.(*errors.ConfigLoadError); ok {
			assert.Greater(t, len(*configErr.HydrationErrors), 0, "should have hydration errors")
			errMsg := configErr.Error()
			assert.Contains(t, errMsg, "hydration errors", "should mention hydration errors")
		}
	})
}

func TestOriginalErrorIssue(t *testing.T) {
	// Test that reproduces the original error issue where empty correlations
	// were being returned instead of nil
	cfg := &AppConfig{}
	mockSource := &mockConfigSource{
		data: map[string]any{
			"name":        "test-app",
			"environment": "dev",
			"version":     "v1.0.0",
			"sub": map[string]any{
				"foo": "test-value",
			},
			"redis": map[string]any{
				"username": "user",
				"password": "pass",
				"timeout":  "5s",
			},
		},
	}

	loader := NewConfigLoader(cfg).Use(mockSource)
	err := loader.Load(context.Background())

	// Should return nil when there are no errors, not an empty CorrelatedError
	assert.NoError(t, err, "should return nil when configuration is successful")

	// The loader's Errors() method should also return nil
	loaderErr := loader.Errors()
	assert.NoError(t, loaderErr, "loader.Errors() should return nil when there are no errors")

	// Make sure we're not getting the old problematic error type
	assert.NotContains(t, fmt.Sprintf("%T", err), "CorrelatedError", "should not return CorrelatedError for successful loads")
}
