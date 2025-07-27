package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "email",
		Value:   "invalid-email",
		Tag:     "email",
		Message: "must be a valid email address",
		Code:    "INVALID_EMAIL",
	}

	assert.Equal(t, "must be a valid email address", err.Error())
}

func TestValidationErrorWithoutMessage(t *testing.T) {
	err := ValidationError{
		Field: "name",
		Value: "",
		Tag:   "required",
	}

	assert.Equal(t, "validation failed for field 'name'", err.Error())
}

func TestFluentErrorBuilder(t *testing.T) {
	tests := []struct {
		name     string
		buildErr func() error
		expected string
	}{
		{
			name: "single field error",
			buildErr: func() error {
				return NewError().
					Field("email").Format("must be a valid email address")
			},
			expected: "must be a valid email address",
		},
		{
			name: "multiple field errors",
			buildErr: func() error {
				return NewError().
					Field("email").Format("must be a valid email address").
					Field("age").Format("must be at least 18")
			},
			expected: "validation failed: must be a valid email address; must be at least 18",
		},
		{
			name: "required field error",
			buildErr: func() error {
				return NewError().
					Field("name").Required()
			},
			expected: "field 'name' is required",
		},
		{
			name: "min length error",
			buildErr: func() error {
				return NewError().
					Field("password").MinLength(8)
			},
			expected: "field 'password' must be at least 8 characters long",
		},
		{
			name: "max length error",
			buildErr: func() error {
				return NewError().
					Field("username").MaxLength(20)
			},
			expected: "field 'username' must be at most 20 characters long",
		},
		{
			name: "range error",
			buildErr: func() error {
				return NewError().
					Field("port").Range(1, 65535)
			},
			expected: "field 'port' must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.buildErr()
			assert.Error(t, err)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestFluentErrorHasErrors(t *testing.T) {
	// Empty error
	err := NewError()
	assert.False(t, err.HasErrors())

	// Add error
	err.Field("email").Format("invalid email")
	assert.True(t, err.HasErrors())
}

func TestFluentErrorErrors(t *testing.T) {
	err := NewError().
		Field("email").Format("invalid email").
		Field("age").Format("too young")

	errors := err.Errors()
	assert.Len(t, errors, 2)
	assert.Equal(t, "email", errors[0].Field)
	assert.Equal(t, "invalid email", errors[0].Message)
	assert.Equal(t, "age", errors[1].Field)
	assert.Equal(t, "too young", errors[1].Message)
}

func TestFluentErrorAdd(t *testing.T) {
	err := NewError()
	validationErr := ValidationError{
		Field:   "test",
		Message: "test error",
	}

	err.Add(validationErr)
	assert.True(t, err.HasErrors())
	assert.Len(t, err.Errors(), 1)
}

func TestFieldErrorBuilderChaining(t *testing.T) {
	err := NewError().
		Field("email").
		Value("invalid@").
		Tag("email").
		Code("INVALID_EMAIL").
		Format("must be a valid email address")

	assert.True(t, err.HasErrors())
	errors := err.Errors()
	assert.Len(t, errors, 1)
	
	validationErr := errors[0]
	assert.Equal(t, "email", validationErr.Field)
	assert.Equal(t, "invalid@", validationErr.Value)
	assert.Equal(t, "email", validationErr.Tag)
	assert.Equal(t, "INVALID_EMAIL", validationErr.Code)
	assert.Equal(t, "must be a valid email address", validationErr.Message)
}

func TestMultiError(t *testing.T) {
	fieldErrors := map[string]error{
		"email": NewError().Field("email").Format("invalid email"),
		"age":   NewError().Field("age").Format("too young"),
	}

	multiErr := NewMultiError(fieldErrors)
	
	assert.True(t, multiErr.HasErrors())
	assert.Len(t, multiErr.Errors(), 2)
	assert.Contains(t, multiErr.Error(), "email:")
	assert.Contains(t, multiErr.Error(), "age:")
}

func TestMultiErrorEmpty(t *testing.T) {
	multiErr := NewMultiError(nil)
	
	assert.False(t, multiErr.HasErrors())
	assert.Equal(t, "", multiErr.Error())
}

func TestMultiErrorAdd(t *testing.T) {
	multiErr := NewMultiError(nil)
	
	err := NewError().Field("test").Format("test error")
	multiErr.Add("test", err)
	
	assert.True(t, multiErr.HasErrors())
	assert.NotNil(t, multiErr.Get("test"))
	assert.Nil(t, multiErr.Get("nonexistent"))
}

func BenchmarkFluentErrorCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewError().
			Field("email").Format("invalid email").
			Field("age").Format("too young").
			Field("name").Required()
	}
}

func BenchmarkMultiErrorCreation(b *testing.B) {
	fieldErrors := map[string]error{
		"email": NewError().Field("email").Format("invalid email"),
		"age":   NewError().Field("age").Format("too young"),
		"name":  NewError().Field("name").Required(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMultiError(fieldErrors)
	}
}