package goconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mateothegreat/go-config/plugins"
	"github.com/stretchr/testify/assert"
)

// --- Mock Sources ---

type StaticSource struct {
	NameStr string
	Data    map[string]any
	Err     error
}

func (s *StaticSource) Name() string {
	return s.NameStr
}

func (s *StaticSource) Load(ctx context.Context) (map[string]any, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Data, nil
}

type MockValidator struct {
	ShouldFail bool
}

func (v MockValidator) Validate(data any) error {
	if v.ShouldFail {
		return errors.New("validation failed")
	}
	return nil
}

// --- Target Config ---

type TestConfig struct {
	Port int    `config:"port"`
	Env  string `config:"env"`
}

// --- Test Cases ---

func TestSingleSource(t *testing.T) {
	cfg := &TestConfig{}
	src := &StaticSource{
		NameStr: "static",
		Data:    map[string]any{"port": 3000, "env": "dev"},
	}

	loader := NewConfigLoader(cfg)
	loader.Use(src)
	err := loader.Load(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "dev", cfg.Env)
}

func TestSourceOrderOverride(t *testing.T) {
	cfg := &TestConfig{}

	low := &StaticSource{
		NameStr: "defaults",
		Data:    map[string]any{"port": 8080, "env": "staging"},
	}
	high := &StaticSource{
		NameStr: "yaml",
		Data:    map[string]any{"port": 9090}, // overrides port only
	}

	loader := NewConfigLoader(cfg)
	loader.Use(low)
	loader.Use(high)
	err := loader.Load(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "staging", cfg.Env)
}

func TestDefaultMerging(t *testing.T) {
	cfg := &TestConfig{}
	main := &StaticSource{
		NameStr: "yaml",
		Data:    map[string]any{"port": 8000},
	}

	defaults := &TestConfig{
		Port: 1234,
		Env:  "default-env",
	}

	loader := NewConfigLoader(cfg)
	loader.Use(main)
	loader.SetDefaults(defaults)
	err := loader.Load(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 8000, cfg.Port)         // from source
	assert.Equal(t, "default-env", cfg.Env) // merged from defaults
}

func TestValidationMockSuccess(t *testing.T) {
	cfg := &TestConfig{}
	src := &StaticSource{
		NameStr: "static",
		Data:    map[string]any{"port": 5000, "env": "qa"},
	}

	loader := NewConfigLoader(cfg)
	loader.Use(src)
	loader.SetValidator(MockValidator{ShouldFail: false})

	err := loader.Load(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 5000, cfg.Port)
}

func TestValidationFailure(t *testing.T) {
	cfg := &TestConfig{}
	src := &StaticSource{
		NameStr: "static",
		Data:    map[string]any{"port": 5000, "env": "qa"},
	}

	loader := NewConfigLoader(cfg)
	loader.Use(src)
	loader.SetValidator(MockValidator{ShouldFail: true})

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestErrorPropagationFromSource(t *testing.T) {
	cfg := &TestConfig{}
	bad := &StaticSource{
		NameStr: "broken",
		Err:     errors.New("read failure"),
	}

	loader := NewConfigLoader(cfg)
	loader.Use(bad)

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read failure")
}

func TestIntrospection(t *testing.T) {
	cfg := &TestConfig{}
	src := &StaticSource{
		NameStr: "introspect",
		Data: map[string]any{
			"port": 9999,
			"env":  "debug",
		},
	}

	loader := NewConfigLoader(cfg)
	loader.Use(src)
	err := loader.Load(context.Background())
	assert.NoError(t, err)

	data := loader.Inspect()
	assert.Equal(t, 9999, data["port"])
	assert.Equal(t, "debug", data["env"])
	names := loader.Sources()
	assert.Contains(t, names, "introspect")
}

func TestMultipleSources(t *testing.T) {

	type c struct {
		Port int    `config:"port"`
		Env  string `config:"env"`
	}

	var conf c

	err := LoadConfig(
		Source("env", plugins.EnvOpts{Prefix: "TEST1"}),
		Source("env", plugins.EnvOpts{Prefix: "TEST2"}),
	).WithDefaults(&c{
		Port: 1234,
		Env:  "default-env",
	}).Build(&conf)

	assert.NoError(t, err)
	assert.Equal(t, 1234, conf.Port)
	assert.Equal(t, "default-env", conf.Env)
}

// Test struct for examples
type SimpleConfig struct {
	Port int    `config:"port" validate:"required,min=1000,max=65535"`
	Env  string `config:"env" validate:"required,oneof=development|staging|production"`
}

func TestExampleSimple(t *testing.T) {
	var conf SimpleConfig

	os.Setenv("PORT", "8080")
	os.Setenv("ENV", "development")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
	}()

	err := LoadConfig(
		Source("env", plugins.EnvOpts{}),
		// Source("yaml", plugins.YAMLOpts{Path: "examples/simple.yaml"}),
	).Build(&conf)

	assert.NoError(t, err)
	assert.Equal(t, 8080, conf.Port)
	assert.Equal(t, "development", conf.Env)
}

func TestDynamicTypeInference(t *testing.T) {
	os.Setenv("TEST_PORT", "8080")
	os.Setenv("TEST_ENV", "production")
	defer func() {
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_ENV")
	}()

	config := &TestConfig{}
	loader := NewConfigLoader(config)

	envPlugin, err := plugins.Env(plugins.EnvOpts{Prefix: "TEST"})
	if err != nil {
		t.Fatal(err)
	}
	loader.Use(envPlugin)

	err = loader.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "production", config.Env)
}

func TestEnvNoPrefix(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("ENV", "production")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
	}()

	config := &TestConfig{}
	loader := NewConfigLoader(config)

	envPlugin, err := plugins.Env(plugins.EnvOpts{}) // No prefix
	if err != nil {
		t.Fatal(err)
	}
	loader.Use(envPlugin)

	err = loader.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "production", config.Env)
}

func TestStringToIntConversion(t *testing.T) {
	// Test explicit string to int conversion
	src := &StaticSource{
		NameStr: "string-conversion",
		Data:    map[string]any{"port": "9000", "env": "test"},
	}

	config := &TestConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 9000, config.Port)
	assert.Equal(t, "test", config.Env)
}

func TestInvalidKeysFiltered(t *testing.T) {
	// Test that keys not present in the struct are filtered out
	src := &StaticSource{
		NameStr: "invalid-keys",
		Data: map[string]any{
			"port":        8080,
			"env":         "test",
			"invalid_key": "should_be_ignored",
			"another_bad": 123,
		},
	}

	config := &TestConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "test", config.Env)

	// Verify that only valid keys are in the merged data
	merged := loader.Inspect()
	assert.Equal(t, 8080, merged["port"])
	assert.Equal(t, "test", merged["env"])
	assert.NotContains(t, merged, "invalid_key")
	assert.NotContains(t, merged, "another_bad")
}

// Validation test structs
type ValidatedConfig struct {
	Port     int    `config:"port" validate:"required,min=1000,max=65535"`
	Env      string `config:"env" validate:"required,oneof=dev|staging|prod"`
	Email    string `config:"email" validate:"required,email"`
	URL      string `config:"url" validate:"required,url"`
	APIKey   string `config:"api_key" validate:"required,len=32,alphanumeric"`
	Password string `config:"password" validate:"required,minlen=8,maxlen=50"`
	Version  string `config:"version" validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
}

func TestValidationSuccess(t *testing.T) {
	src := &StaticSource{
		NameStr: "valid-config",
		Data: map[string]any{
			"port":     8080,
			"env":      "dev",
			"email":    "test@example.com",
			"url":      "https://example.com",
			"api_key":  "abc123def456ghi789jkl012mno345pq",
			"password": "securepassword123",
			"version":  "v1.2.3",
		},
	}

	config := &ValidatedConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "dev", config.Env)
}

func TestValidationRequired(t *testing.T) {
	src := &StaticSource{
		NameStr: "missing-required",
		Data: map[string]any{
			"env": "dev", // missing required port
		},
	}

	config := &ValidatedConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Port' is required")
}

func TestValidationMinMax(t *testing.T) {
	src := &StaticSource{
		NameStr: "invalid-port",
		Data: map[string]any{
			"port": 999, // below minimum
			"env":  "dev",
		},
	}

	config := &ValidatedConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Port' must be at least 1000")
}

func TestValidationOneOf(t *testing.T) {
	src := &StaticSource{
		NameStr: "invalid-env",
		Data: map[string]any{
			"port": 8080,
			"env":  "invalid", // not in allowed values
		},
	}

	config := &ValidatedConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Env' must be one of")
}

func TestValidationEmail(t *testing.T) {
	src := &StaticSource{
		NameStr: "invalid-email",
		Data: map[string]any{
			"port":  8080,
			"env":   "dev",
			"email": "not-an-email",
		},
	}

	config := &ValidatedConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Email' must be a valid email address")
}

func TestValidationRegex(t *testing.T) {
	src := &StaticSource{
		NameStr: "invalid-version",
		Data: map[string]any{
			"port":    8080,
			"env":     "dev",
			"version": "1.2.3", // missing 'v' prefix
		},
	}

	config := &ValidatedConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Version' does not match pattern")
}

func TestCustomValidator(t *testing.T) {
	type CustomConfig struct {
		Name string `config:"name" validate:"custom_uppercase"`
	}

	src := &StaticSource{
		NameStr: "custom-validation",
		Data: map[string]any{
			"name": "lowercase",
		},
	}

	config := &CustomConfig{}
	loader := NewConfigLoader(config)
	loader.Use(src)

	// Add custom validator that requires uppercase
	loader.SetStructValidator(func(v *StructValidator) {
		v.WithCustomValidator("custom_uppercase", func(fieldName string, value reflect.Value, rule string) error {
			str := value.String()
			if str != strings.ToUpper(str) {
				return fmt.Errorf("field '%s' must be uppercase", fieldName)
			}
			return nil
		})
	})

	err := loader.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Name' must be uppercase")
}

// Test the new LoadWithPlugins API
func TestLoadWithPluginsAPI(t *testing.T) {
	os.Setenv("TEST_PORT", "9000")
	os.Setenv("TEST_ENV", "staging")
	defer func() {
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_ENV")
	}()

	var config TestConfig

	err := LoadWithPlugins(Env(plugins.EnvOpts{Prefix: "TEST"})).Build(&config)
	assert.NoError(t, err)
	assert.Equal(t, 9000, config.Port)
	assert.Equal(t, "staging", config.Env)
}

func TestLoadWithPluginsWithDefaults(t *testing.T) {
	var config TestConfig

	defaults := &TestConfig{
		Port: 3000,
		Env:  "default",
	}

	err := LoadWithPlugins(Env(plugins.EnvOpts{Prefix: "MISSING"})).
		WithDefaults(defaults).
		Build(&config)

	assert.NoError(t, err)
	assert.Equal(t, 3000, config.Port)
	assert.Equal(t, "default", config.Env)
}

func TestLoadWithPluginsWithValidation(t *testing.T) {
	var config ValidatedConfig

	// Set environment variables that will fail validation
	os.Setenv("PORT", "999") // Below minimum
	os.Setenv("ENV", "dev")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
	}()

	err := LoadWithPlugins(Env(plugins.EnvOpts{})).Build(&config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Port' must be at least 1000")
}

func TestLoadWithPluginsCustomValidator(t *testing.T) {
	type CustomConfig struct {
		Name string `config:"name" validate:"uppercase"`
	}

	// Set environment variable
	os.Setenv("NAME", "lowercase")
	defer os.Unsetenv("NAME")

	var config CustomConfig

	err := LoadWithPlugins(Env(plugins.EnvOpts{})).
		WithStructValidator(func(v *StructValidator) {
			v.WithCustomValidator("uppercase", func(fieldName string, value reflect.Value, _ string) error {
				str := value.String()
				if str != strings.ToUpper(str) {
					return fmt.Errorf("field '%s' must be uppercase", fieldName)
				}
				return nil
			})
		}).
		Build(&config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Name' must be uppercase")
}

func TestLoadWithMultiplePlugins(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `port: 8080
env: development`

	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Set env vars that would override
	os.Setenv("PORT", "9000")
	defer os.Unsetenv("PORT")

	var config SimpleConfig

	// YAML plugin first, then env (env should override)
	err = LoadWithPlugins(
		YAML(plugins.YAMLOpts{Path: tmpFile.Name()}),
		Env(plugins.EnvOpts{}),
	).Build(&config)
	assert.NoError(t, err)
	assert.Equal(t, 9000, config.Port)         // From env override
	assert.Equal(t, "development", config.Env) // From YAML
}

func TestValidationErrorsType(t *testing.T) {
	type TestValidationConfig struct {
		Port  int    `config:"port" validate:"required,min=1000"`
		Email string `config:"email" validate:"required,email"`
	}

	// Set invalid values
	os.Setenv("PORT", "999")      // Below minimum
	os.Setenv("EMAIL", "invalid") // Invalid email
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("EMAIL")
	}()

	var config TestValidationConfig

	err := LoadWithPlugins(Env(plugins.EnvOpts{})).Build(&config)
	assert.Error(t, err)

	// Test that we can access individual validation errors
	if validationErrors, ok := AsValidationErrors(err); ok {
		errors := validationErrors.Errors()
		assert.Len(t, errors, 2)
		
		// Check that we have the expected error messages
		foundPortError := false
		foundEmailError := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "Port") && strings.Contains(errMsg, "must be at least 1000") {
				foundPortError = true
			}
			if strings.Contains(errMsg, "Email") && strings.Contains(errMsg, "valid email") {
				foundEmailError = true
			}
		}
		assert.True(t, foundPortError, "Expected to find port validation error")
		assert.True(t, foundEmailError, "Expected to find email validation error")
		
		assert.True(t, validationErrors.HasErrors())
	} else {
		t.Fatal("Expected ValidationErrors type")
	}
}

func TestFastStructValidator(t *testing.T) {
	type TestConfig struct {
		Name  string `validate:"required,minlen=3"`
		Age   int    `validate:"min=18,max=99"`
		Email string `validate:"required,email"`
	}

	// Test valid config
	validConfig := &TestConfig{
		Name:  "John",
		Age:   25,
		Email: "john@example.com",
	}

	fastValidator := NewFastStructValidator()
	err := fastValidator.Validate(validConfig)
	assert.NoError(t, err)

	// Test invalid config
	invalidConfig := &TestConfig{
		Name:  "Jo",            // Too short
		Age:   17,             // Too young
		Email: "invalid-email", // Invalid email
	}

	err = fastValidator.Validate(invalidConfig)
	assert.Error(t, err)

	if validationErrors, ok := AsValidationErrors(err); ok {
		errors := validationErrors.Errors()
		assert.Len(t, errors, 3)
		
		// Check for specific error messages
		errorString := strings.Join(errors, " ")
		assert.Contains(t, errorString, "must be at least 3 characters long")
		assert.Contains(t, errorString, "must be at least 18")
		assert.Contains(t, errorString, "valid email address")
	} else {
		t.Fatal("Expected ValidationErrors type")
	}
}
