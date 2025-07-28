package main

import (
	"reflect"
	"strings"
	"testing"
)

// BenchmarkGeneratedValidation tests the performance of generated validation.
func BenchmarkGeneratedValidation(b *testing.B) {
	config := &ServerConfig{
		Name:     "test-server",
		Port:     8080,
		Host:     "localhost",
		Email:    "test@example.com",
		LogLevel: "info",
		Debug:    false,
		Version:  "v1.0.0",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := config.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReflectionValidation simulates reflection-based validation for comparison.
func BenchmarkReflectionValidation(b *testing.B) {
	config := &ServerConfig{
		Name:     "test-server",
		Port:     8080,
		Host:     "localhost",
		Email:    "test@example.com",
		LogLevel: "info",
		Debug:    false,
		Version:  "v1.0.0",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := reflectionValidate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// reflectionValidate simulates how reflection-based validators work.
func reflectionValidate(config interface{}) error {
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		tag := field.Tag.Get("validate")

		if tag == "" {
			continue
		}

		// Simulate tag parsing (allocates strings).
		rules := strings.Split(tag, ",")

		// Simulate validation logic with reflection calls.
		for _, rule := range rules {
			if strings.HasPrefix(rule, "required") {
				switch value.Kind() {
				case reflect.String:
					if value.String() == "" {
						return reflect.ValueOf("required field is empty").Interface().(error)
					}
				case reflect.Int:
					if value.Int() == 0 {
						return reflect.ValueOf("required field is zero").Interface().(error)
					}
				}
			}

			// More rule checking would happen here...
			// Each reflection call adds overhead.
		}
	}

	return nil
}
