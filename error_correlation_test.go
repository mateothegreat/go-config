package goconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	errorspkg "github.com/mateothegreat/go-config/errors"
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig for error correlation testing
type ErrorTestConfig struct {
	Name    string `validate:"required,minlen=3" yaml:"name"`
	Port    int    `validate:"min=1000,max=65535" yaml:"port"`
	Email   string `validate:"required,email" yaml:"email"`
	Timeout int    `validate:"min=1,max=300" yaml:"timeout"`
}

func TestErrorCorrelationMissingFile(t *testing.T) {
	config := &ErrorTestConfig{}

	// Create a loader with a non-existent YAML source
	yamlPlugin, err := CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "non-existent.yaml"})
	require.NoError(t, err)

	loader := NewLoader(config)
	loader.Use(yamlPlugin)
	err = loader.Load(context.Background())
	require.Error(t, err)

	// Should get correlated error
	if correlatedErr, ok := err.(*CorrelatedError); ok {
		errorMsg := strings.ToLower(correlatedErr.Error())
		suggestions := strings.ToLower(strings.Join(correlatedErr.GetSuggestions(), " "))

		// Should contain file-related errors and suggestions
		assert.True(t, strings.Contains(errorMsg, "file") || strings.Contains(suggestions, "file"))
		assert.True(t, strings.Contains(errorMsg, "yaml") || strings.Contains(suggestions, "yaml"))
	} else {
		t.Error("Expected CorrelatedError but got different error type")
	}
}

func TestErrorCorrelationSourceAndValidation(t *testing.T) {
	// Create temporary YAML file with invalid content
	tmpFile, err := os.CreateTemp("", "test_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write invalid YAML content that will pass parsing but fail validation
	yamlContent := `name: ab
port: 70000
email: invalid-email
timeout: -1
`
	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	config := &ErrorTestConfig{}

	// Load with the invalid YAML
	yamlPlugin, err := CreateSourcePlugin("yaml", sources.YAMLOpts{Path: tmpFile.Name()})
	require.NoError(t, err)

	loader := NewLoader(config)
	loader.Use(yamlPlugin)
	err = loader.Load(context.Background())
	require.Error(t, err)

	// Should get correlated error with both source and validation info
	if correlatedErr, ok := err.(*CorrelatedError); ok {
		errorMsg := strings.ToLower(correlatedErr.Error())
		suggestions := strings.ToLower(strings.Join(correlatedErr.GetSuggestions(), " "))

		// Check for validation-related content
		assert.True(t,
			strings.Contains(errorMsg, "validation") ||
				strings.Contains(suggestions, "validation") ||
				strings.Contains(errorMsg, "invalid") ||
				strings.Contains(suggestions, "check"))

		// Check for field-specific content
		assert.True(t,
			strings.Contains(errorMsg, "name") || strings.Contains(suggestions, "name") ||
				strings.Contains(errorMsg, "port") || strings.Contains(suggestions, "port") ||
				strings.Contains(errorMsg, "email") || strings.Contains(suggestions, "email"))

		// Should have suggestions
		assert.True(t,
			strings.Contains(suggestions, "check") || strings.Contains(suggestions, "verify") ||
				strings.Contains(suggestions, "fix") || strings.Contains(suggestions, "ensure"))
	} else {
		t.Error("Expected CorrelatedError but got different error type")
	}
}

func TestErrorCorrelationCategories(t *testing.T) {
	testCases := []struct {
		name     string
		config   ErrorTestConfig
		expected []string
	}{
		{
			name: "missing_required",
			config: ErrorTestConfig{
				Name:    "", // Empty name (required)
				Port:    8080,
				Email:   "test@example.com",
				Timeout: 30,
			},
			expected: []string{"required", "name"},
		},
		{
			name: "format_error",
			config: ErrorTestConfig{
				Name:    "test",
				Port:    8080,
				Email:   "invalid-email", // Invalid email format
				Timeout: 30,
			},
			expected: []string{"email", "valid"},
		},
		{
			name: "range_error",
			config: ErrorTestConfig{
				Name:    "test",
				Port:    70000, // Port too high
				Email:   "test@example.com",
				Timeout: 30,
			},
			expected: []string{"port", "65535"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &tc.config

			// Create temp file with test data to trigger validation
			tmpFile, err := os.CreateTemp("", "test_correlation_*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			yamlContent := fmt.Sprintf(`name: %s
port: %d
email: %s
timeout: %d
`, config.Name, config.Port, config.Email, config.Timeout)
			_, err = tmpFile.WriteString(yamlContent)
			require.NoError(t, err)
			tmpFile.Close()

			yamlPlugin, err := CreateSourcePlugin("yaml", sources.YAMLOpts{Path: tmpFile.Name()})
			require.NoError(t, err)

			loader := NewLoader(&ErrorTestConfig{})
			loader.Use(yamlPlugin)
			err = loader.Load(context.Background())

			require.Error(t, err)

			// Should get correlated error
			if correlatedErr, ok := err.(*CorrelatedError); ok {
				errorMsg := strings.ToLower(correlatedErr.Error())
				suggestions := strings.ToLower(strings.Join(correlatedErr.GetSuggestions(), " "))

				// Check that expected terms appear in error or suggestions
				for _, expected := range tc.expected {
					assert.True(t,
						strings.Contains(errorMsg, expected) ||
							strings.Contains(suggestions, expected),
						"Expected '%s' to appear in error message or suggestions. Error: %s, Suggestions: %v",
						expected, correlatedErr.Error(), correlatedErr.GetSuggestions())
				}
			} else {
				t.Errorf("Expected CorrelatedError but got: %T - %v", err, err)
			}
		})
	}
}

func TestErrorCorrelationSuggestionQuality(t *testing.T) {
	// Test that suggestions are actually helpful
	tmpFile, err := os.CreateTemp("", "test_suggestions_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	yamlContent := `name: ab
port: 70000
email: invalid
timeout: -1
`
	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	yamlPlugin, err := CreateSourcePlugin("yaml", sources.YAMLOpts{Path: tmpFile.Name()})
	require.NoError(t, err)

	loader := NewLoader(&ErrorTestConfig{})
	loader.Use(yamlPlugin)
	err = loader.Load(context.Background())

	require.Error(t, err)

	if correlatedErr, ok := err.(*CorrelatedError); ok {
		suggestions := correlatedErr.GetSuggestions()

		// Should have suggestions
		assert.NotEmpty(t, suggestions, "Expected helpful suggestions")

		// Suggestions should be actionable (contain action words)
		suggestionText := strings.ToLower(strings.Join(suggestions, " "))
		hasActionableContent := strings.Contains(suggestionText, "check") ||
			strings.Contains(suggestionText, "verify") ||
			strings.Contains(suggestionText, "ensure") ||
			strings.Contains(suggestionText, "fix") ||
			strings.Contains(suggestionText, "set") ||
			strings.Contains(suggestionText, "change")

		assert.True(t, hasActionableContent, "Suggestions should contain actionable advice: %v", suggestions)
	} else {
		t.Errorf("Expected CorrelatedError but got: %T - %v", err, err)
	}
}

func TestErrorCorrelatorDirectUsage(t *testing.T) {
	// Test direct usage of ErrorCorrelator
	correlator := NewErrorCorrelator()

	// Add a source error
	sourceErr := errors.New("failed to read config.yaml")
	correlator.AddSourceError("yaml", sourceErr)

	// Add validation errors
	validationErrs := NewError().Field("Name").Required().Field("Email").Format("invalid format")
	correlator.AddValidationErrors(validationErrs.ValidationErrors())

	// Create correlated error
	correlations := correlator.Correlate()
	summary := correlator.GetCorrelationSummary()
	correlatedErr := errorspkg.NewCorrelatedError(correlations, summary)

	// Should be of correct type
	assert.IsType(t, &CorrelatedError{}, correlatedErr)

	// Should have suggestions
	suggestions := correlatedErr.GetSuggestions()
	assert.NotEmpty(t, suggestions)

	// Should contain relevant info
	errorMsg := correlatedErr.Error()
	assert.True(t, len(errorMsg) > 0, "Error message should not be empty")
}

func TestCorrelatedErrorInterface(t *testing.T) {
	// Test that CorrelatedError properly implements error interface
	correlator := NewErrorCorrelator()
	correlator.AddSourceError("test", errors.New("test error"))

	correlations := correlator.Correlate()
	summary := correlator.GetCorrelationSummary()
	correlatedErr := errorspkg.NewCorrelatedError(correlations, summary)

	// Should implement error interface
	var err error = correlatedErr
	assert.NotEmpty(t, err.Error())

	// Should be identifiable as CorrelatedError
	if ce, ok := err.(*CorrelatedError); ok {
		assert.NotEmpty(t, ce.GetSuggestions())
	} else {
		t.Error("CorrelatedError should be identifiable via type assertion")
	}
}
