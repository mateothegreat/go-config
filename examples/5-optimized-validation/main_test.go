package main

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/mateothegreat/go-config/validation"
)

func TestOptimizedValidation(t *testing.T) {
	t.Run("ServerConfig_Valid", func(t *testing.T) {
		config := &ServerConfig{
			Name:        "api-server",
			Environment: "prod",
			Host:        "api.example.com",
			Port:        8080,
			Email:       "admin@example.com",
			AdminEmails: []string{"admin@example.com", "support@example.com"},
			Debug:       true,
			Enabled:     true,
		}

		errors := config.Validate()
		if errors.HasErrors() {
			t.Errorf("Expected no errors for valid config, got: %v", errors)
		}

		result := config.ValidateWithResult()
		if result.HasErrors() {
			t.Errorf("Expected no errors in result for valid config, got: %v", result.Error())
		}
	})

	t.Run("ServerConfig_Invalid", func(t *testing.T) {
		config := &ServerConfig{
			Name:        "x",              // Too short
			Environment: "invalid",       // Not in oneof
			Host:        "invalid_host",  // Invalid hostname
			Port:        0,               // Required, zero value
			Email:       "invalid-email", // Invalid email
			AdminEmails: []string{},      // Required, empty
			Debug:       false,           // Required bool
		}

		errors := config.Validate()
		if !errors.HasErrors() {
			t.Error("Expected errors for invalid config")
		}

		// Check specific errors
		if !errors.HasError(validation.ErrMinLength) {
			t.Error("Expected min length error for Name")
		}
		if !errors.HasError(validation.ErrOneOf) {
			t.Error("Expected oneof error for Environment")
		}
		if !errors.HasError(validation.ErrHostname) {
			t.Error("Expected hostname error for Host")
		}
		if !errors.HasError(validation.ErrRequired) {
			t.Error("Expected required error for Port and AdminEmails")
		}
		if !errors.HasError(validation.ErrEmail) {
			t.Error("Expected email error for Email")
		}

		result := config.ValidateWithResult()
		if !result.HasErrors() {
			t.Error("Expected errors in result for invalid config")
		}

		// Check field-specific errors
		if !result.GetFieldErrors("Name").HasError(validation.ErrMinLength) {
			t.Error("Expected min length error for Name field")
		}
		if !result.GetFieldErrors("Environment").HasError(validation.ErrOneOf) {
			t.Error("Expected oneof error for Environment field")
		}
	})

	t.Run("DatabaseConfig_Valid", func(t *testing.T) {
		config := &DatabaseConfig{
			Driver:       "postgres",
			Host:         "db.example.com",
			Port:         5432,
			Username:     "dbuser",
			Password:     "dbpass",
			Database:     "myapp",
			MaxOpenConns: 25,
			MaxIdleConns: 10,
		}

		errors := config.Validate()
		if errors.HasErrors() {
			t.Errorf("Expected no errors for valid database config, got: %v", errors)
		}
	})

	t.Run("APIConfig_Valid", func(t *testing.T) {
		config := &APIConfig{
			BaseURL:   "https://api.example.com",
			APIKey:    "SGVsbG8gV29ybGQ=", // Valid base64
			Version:   "v1.0.0",
			Timeout:   30.5,
			RateLimit: 1000,
		}

		errors := config.Validate()
		if errors.HasErrors() {
			t.Errorf("Expected no errors for valid API config, got: %v", errors)
		}
	})

	t.Run("Config_NestedValidation", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Name:        "api-server",
				Environment: "prod",
				Host:        "api.example.com",
				Port:        8080,
				Email:       "admin@example.com",
				AdminEmails: []string{"admin@example.com"},
				Debug:       true,
			},
			Database: DatabaseConfig{
				Driver:       "postgres",
				Host:         "db.example.com",
				Port:         5432,
				Username:     "user",
				Password:     "pass",
				Database:     "db",
				MaxOpenConns: 25,
				MaxIdleConns: 10,
			},
			API: APIConfig{
				BaseURL:   "https://api.example.com",
				APIKey:    "SGVsbG8gV29ybGQ=",
				Version:   "v1.0.0",
				Timeout:   30.0,
				RateLimit: 1000,
			},
			LogLevel: "info",
			Debug:    false,
		}

		errors := config.Validate()
		if errors.HasErrors() {
			t.Errorf("Expected no errors for valid nested config, got: %v", errors)
		}

		result := config.ValidateWithResult()
		if result.HasErrors() {
			t.Errorf("Expected no errors in nested result, got: %v", result.Error())
		}
	})

	t.Run("Config_NestedValidationErrors", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Name:        "", // Invalid: required
				Environment: "invalid",
				Host:        "api.example.com",
				Port:        8080,
				Email:       "admin@example.com",
				AdminEmails: []string{"admin@example.com"},
				Debug:       true,
			},
			Database: DatabaseConfig{
				Driver: "", // Invalid: required
				Host:   "db.example.com",
				Port:   5432,
			},
			API: APIConfig{
				BaseURL: "", // Invalid: required
				APIKey:  "SGVsbG8gV29ybGQ=",
				Version: "v1.0.0",
			},
			LogLevel: "invalid", // Invalid: not in oneof
		}

		result := config.ValidateWithResult()
		if !result.HasErrors() {
			t.Error("Expected errors for invalid nested config")
		}

		// Check nested field errors
		if !result.GetFieldErrors("Server.Name").HasError(validation.ErrRequired) {
			t.Error("Expected required error for Server.Name")
		}
		if !result.GetFieldErrors("Database.Driver").HasError(validation.ErrRequired) {
			t.Error("Expected required error for Database.Driver")
		}
		if !result.GetFieldErrors("API.BaseURL").HasError(validation.ErrRequired) {
			t.Error("Expected required error for API.BaseURL")
		}
		if !result.GetFieldErrors("LogLevel").HasError(validation.ErrOneOf) {
			t.Error("Expected oneof error for LogLevel")
		}
	})
}

// Benchmark the optimized validation vs reflection-based validation
func BenchmarkOptimizedValidation(b *testing.B) {
	validConfig := &ServerConfig{
		Name:        "api-server",
		Environment: "prod",
		Host:        "api.example.com",
		Port:        8080,
		Email:       "admin@example.com",
		AdminEmails: []string{"admin@example.com", "support@example.com"},
		Debug:       true,
		Enabled:     true,
	}

	b.Run("OptimizedValidation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errors := validConfig.Validate()
			_ = errors.HasErrors()
		}
	})

	b.Run("OptimizedValidationWithResult", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := validConfig.ValidateWithResult()
			_ = result.HasErrors()
		}
	})

	b.Run("ReflectionValidation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errors := validateWithReflection(validConfig)
			_ = len(errors) > 0
		}
	})
}

// Benchmark nested struct validation
func BenchmarkNestedValidation(b *testing.B) {
	config := &Config{
		Server: ServerConfig{
			Name:        "api-server",
			Environment: "prod",
			Host:        "api.example.com",
			Port:        8080,
			Email:       "admin@example.com",
			AdminEmails: []string{"admin@example.com"},
			Debug:       true,
		},
		Database: DatabaseConfig{
			Driver:   "postgres",
			Host:     "db.example.com",
			Port:     5432,
			Username: "user",
			Password: "pass",
			Database: "db",
		},
		API: APIConfig{
			BaseURL:   "https://api.example.com",
			APIKey:    "SGVsbG8gV29ybGQ=",
			Version:   "v1.0.0",
			Timeout:   30.0,
			RateLimit: 1000,
		},
		LogLevel: "info",
	}

	b.Run("OptimizedNested", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errors := config.Validate()
			_ = errors.HasErrors()
		}
	})

	b.Run("OptimizedNestedWithResult", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := config.ValidateWithResult()
			_ = result.HasErrors()
		}
	})

	b.Run("ReflectionNested", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errors := validateNestedWithReflection(config)
			_ = len(errors) > 0
		}
	})
}

// Benchmark error handling approaches
func BenchmarkErrorHandling(b *testing.B) {
	b.Run("BitFieldErrors", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var errors validation.ValidationErrors
			errors |= validation.ErrRequired
			errors |= validation.ErrEmail
			errors |= validation.ErrMinLength
			_ = errors.HasErrors()
			_ = errors.Count()
			_ = errors.Error()
		}
	})

	b.Run("SliceErrors", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var errors []string
			errors = append(errors, "required")
			errors = append(errors, "email")
			errors = append(errors, "minlength")
			_ = len(errors) > 0
			_ = len(errors)
			_ = strings.Join(errors, ", ")
		}
	})
}

// validateWithReflection simulates reflection-based validation for benchmarking
func validateWithReflection(obj interface{}) []string {
	var errors []string
	val := reflect.ValueOf(obj).Elem() // Assume pointer to struct
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		tag := field.Tag.Get("validate")

		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			if strings.Contains(rule, "=") {
				parts := strings.SplitN(rule, "=", 2)
				ruleName := parts[0]
				ruleValue := parts[1]

				switch ruleName {
				case "minlen":
					if value.Kind() == reflect.String {
						minLen, _ := strconv.Atoi(ruleValue)
						if len(value.String()) < minLen {
							errors = append(errors, field.Name+": minlen")
						}
					}
				case "maxlen":
					if value.Kind() == reflect.String {
						maxLen, _ := strconv.Atoi(ruleValue)
						if len(value.String()) > maxLen {
							errors = append(errors, field.Name+": maxlen")
						}
					}
				case "min":
					switch value.Kind() {
					case reflect.Int:
						min, _ := strconv.Atoi(ruleValue)
						if int(value.Int()) < min {
							errors = append(errors, field.Name+": min")
						}
					case reflect.Float32, reflect.Float64:
						min, _ := strconv.ParseFloat(ruleValue, 64)
						if value.Float() < min {
							errors = append(errors, field.Name+": min")
						}
					}
				case "max":
					switch value.Kind() {
					case reflect.Int:
						max, _ := strconv.Atoi(ruleValue)
						if int(value.Int()) > max {
							errors = append(errors, field.Name+": max")
						}
					case reflect.Float32, reflect.Float64:
						max, _ := strconv.ParseFloat(ruleValue, 64)
						if value.Float() > max {
							errors = append(errors, field.Name+": max")
						}
					}
				case "oneof":
					if value.Kind() == reflect.String {
						options := strings.Split(ruleValue, " ")
						found := false
						for _, opt := range options {
							if value.String() == opt {
								found = true
								break
							}
						}
						if !found {
							errors = append(errors, field.Name+": oneof")
						}
					}
				}
			} else {
				switch rule {
				case "required":
					switch value.Kind() {
					case reflect.String:
						if value.String() == "" {
							errors = append(errors, field.Name+": required")
						}
					case reflect.Int:
						if value.Int() == 0 {
							errors = append(errors, field.Name+": required")
						}
					case reflect.Bool:
						if !value.Bool() {
							errors = append(errors, field.Name+": required")
						}
					case reflect.Slice:
						if value.Len() == 0 {
							errors = append(errors, field.Name+": required")
						}
					}
				case "email":
					if value.Kind() == reflect.String {
						email := value.String()
						if email != "" && !strings.Contains(email, "@") {
							errors = append(errors, field.Name+": email")
						}
					}
				case "url":
					if value.Kind() == reflect.String {
						url := value.String()
						if url != "" && !strings.HasPrefix(url, "http") {
							errors = append(errors, field.Name+": url")
						}
					}
				case "hostname":
					if value.Kind() == reflect.String {
						hostname := value.String()
						if hostname != "" && strings.Contains(hostname, "_") {
							errors = append(errors, field.Name+": hostname")
						}
					}
				case "base64":
					if value.Kind() == reflect.String {
						b64 := value.String()
						if b64 != "" && len(b64)%4 != 0 {
							errors = append(errors, field.Name+": base64")
						}
					}
				}
			}
		}
	}

	return errors
}

// validateNestedWithReflection simulates nested reflection-based validation
func validateNestedWithReflection(obj interface{}) []string {
	var errors []string
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		if value.Kind() == reflect.Struct {
			// Recursively validate nested struct
			nestedErrors := validateWithReflection(value.Addr().Interface())
			for _, err := range nestedErrors {
				errors = append(errors, field.Name+"."+err)
			}
		} else {
			// Regular field validation
			fieldErrors := validateWithReflection(reflect.New(reflect.StructOf([]reflect.StructField{field})).Interface())
			errors = append(errors, fieldErrors...)
		}
	}

	return errors
}