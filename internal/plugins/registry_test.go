package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test plugin implementation
type TestValidator struct {
	validateFunc func(interface{}) error
}

func (tv *TestValidator) Validate(field interface{}) error {
	if tv.validateFunc != nil {
		return tv.validateFunc(field)
	}
	return nil
}

func TestNewPluginRegistry(t *testing.T) {
	registry := NewPluginRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.plugins)
	assert.NotNil(t, registry.metadata)
}

func TestRegisterValidator(t *testing.T) {
	registry := NewPluginRegistry()
	validator := &TestValidator{}
	
	registry.RegisterValidator("test", validator)
	
	// Check that validator was registered
	retrieved, exists := registry.GetValidator("test")
	assert.True(t, exists)
	assert.Equal(t, validator, retrieved)
}

func TestRegisterValidatorWithMetadata(t *testing.T) {
	registry := NewPluginRegistry()
	validator := &TestValidator{}
	metadata := PluginMetadata{
		Name:        "Test Validator",
		Version:     "1.0.0",
		Description: "A test validator",
		Author:      "Test Author",
		SupportedTypes: []string{"string"},
	}
	
	registry.RegisterValidatorWithMetadata("test", validator, metadata)
	
	// Check validator
	retrieved, exists := registry.GetValidator("test")
	assert.True(t, exists)
	assert.Equal(t, validator, retrieved)
	
	// Check metadata
	retrievedMetadata, exists := registry.GetMetadata("test")
	assert.True(t, exists)
	assert.Equal(t, metadata, retrievedMetadata)
}

func TestGetValidatorNotFound(t *testing.T) {
	registry := NewPluginRegistry()
	
	validator, exists := registry.GetValidator("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, validator)
}

func TestListValidators(t *testing.T) {
	registry := NewPluginRegistry()
	
	// Initially empty
	validators := registry.ListValidators()
	assert.Empty(t, validators)
	
	// Add some validators
	registry.RegisterValidator("test1", &TestValidator{})
	registry.RegisterValidator("test2", &TestValidator{})
	
	validators = registry.ListValidators()
	assert.Len(t, validators, 2)
	assert.Contains(t, validators, "test1")
	assert.Contains(t, validators, "test2")
}

func TestUnregisterValidator(t *testing.T) {
	registry := NewPluginRegistry()
	validator := &TestValidator{}
	metadata := PluginMetadata{Name: "Test"}
	
	registry.RegisterValidatorWithMetadata("test", validator, metadata)
	
	// Verify it exists
	_, exists := registry.GetValidator("test")
	assert.True(t, exists)
	
	// Unregister
	registry.UnregisterValidator("test")
	
	// Verify it's gone
	_, exists = registry.GetValidator("test")
	assert.False(t, exists)
	
	// Metadata should also be gone
	_, exists = registry.GetMetadata("test")
	assert.False(t, exists)
}

func TestClear(t *testing.T) {
	registry := NewPluginRegistry()
	
	// Add some validators
	registry.RegisterValidator("test1", &TestValidator{})
	registry.RegisterValidator("test2", &TestValidator{})
	
	assert.Len(t, registry.ListValidators(), 2)
	
	// Clear
	registry.Clear()
	
	// Should be empty
	assert.Empty(t, registry.ListValidators())
}

func TestValidateWithPlugin(t *testing.T) {
	registry := NewPluginRegistry()
	
	// Register a validator that always returns an error
	validator := &TestValidator{
		validateFunc: func(field interface{}) error {
			return assert.AnError
		},
	}
	registry.RegisterValidator("test", validator)
	
	// Test validation
	err := registry.ValidateWithPlugin("test", "some value")
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestValidateWithPluginNotFound(t *testing.T) {
	registry := NewPluginRegistry()
	
	err := registry.ValidateWithPlugin("nonexistent", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validator plugin 'nonexistent' not found")
}

func TestGlobalRegistry(t *testing.T) {
	// Test global registry functions
	validator := &TestValidator{}
	
	RegisterValidator("global_test", validator)
	
	retrieved, exists := GetValidator("global_test")
	assert.True(t, exists)
	assert.Equal(t, validator, retrieved)
	
	validators := ListValidators()
	assert.Contains(t, validators, "global_test")
}

func TestBuiltInValidators(t *testing.T) {
	tests := []struct {
		name      string
		validator string
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "valid IP",
			validator: "ip",
			value:     "192.168.1.1",
			wantErr:   false,
		},
		{
			name:      "invalid IP type",
			validator: "ip",
			value:     123,
			wantErr:   true,
		},
		{
			name:      "valid UUID",
			validator: "uuid",
			value:     "123e4567-e89b-12d3-a456-426614174000",
			wantErr:   false,
		},
		{
			name:      "invalid UUID",
			validator: "uuid",
			value:     "not-a-uuid",
			wantErr:   true,
		},
		{
			name:      "valid credit card",
			validator: "creditcard",
			value:     "4111111111111111",
			wantErr:   false,
		},
		{
			name:      "invalid credit card",
			validator: "creditcard",
			value:     "123",
			wantErr:   true,
		},
		{
			name:      "valid phone",
			validator: "phone",
			value:     "1234567890",
			wantErr:   false,
		},
		{
			name:      "invalid phone",
			validator: "phone",
			value:     "123",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, exists := GetValidator(tt.validator)
			assert.True(t, exists, "Built-in validator %s should exist", tt.validator)
			
			err := validator.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuiltInValidatorMetadata(t *testing.T) {
	registry := globalRegistry
	
	builtInValidators := []string{"ip", "uuid", "creditcard", "phone"}
	
	for _, validatorName := range builtInValidators {
		t.Run(validatorName, func(t *testing.T) {
			metadata, exists := registry.GetMetadata(validatorName)
			assert.True(t, exists)
			assert.NotEmpty(t, metadata.Name)
			assert.NotEmpty(t, metadata.Version)
			assert.NotEmpty(t, metadata.Description)
			assert.Equal(t, "go-validate", metadata.Author)
			assert.Contains(t, metadata.SupportedTypes, "string")
		})
	}
}

func BenchmarkValidatorRegistration(b *testing.B) {
	registry := NewPluginRegistry()
	validator := &TestValidator{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "test"
		registry.RegisterValidator(name, validator)
		registry.UnregisterValidator(name)
	}
}

func BenchmarkValidatorLookup(b *testing.B) {
	registry := NewPluginRegistry()
	validator := &TestValidator{}
	registry.RegisterValidator("test", validator)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetValidator("test")
	}
}

func BenchmarkValidateWithPlugin(b *testing.B) {
	registry := NewPluginRegistry()
	validator := &TestValidator{
		validateFunc: func(field interface{}) error {
			return nil // No error for benchmark
		},
	}
	registry.RegisterValidator("test", validator)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.ValidateWithPlugin("test", "test value")
	}
}