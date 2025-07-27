package plugins

import (
	"context"
	"os"
	"strconv"
	"strings"
)

// EnvOpts are the options for the env plugin.
type EnvOpts struct {
	Prefix string
}

// EnvSource is the source for the env plugin.
type EnvSource struct {
	Prefix string
}

// Env creates a new env plugin.
func Env(opts any) (Plugin, error) {
	o := opts.(EnvOpts)
	return &EnvSource{Prefix: o.Prefix}, nil
}

// Name returns the name of the plugin.
func (e *EnvSource) Name() string {
	return "env"
}

// Load loads the plugin.
func (e *EnvSource) Load(ctx context.Context) (map[string]any, error) {
	result := make(map[string]any)
	prefix := strings.ToUpper(e.Prefix)
	if !strings.HasSuffix(prefix, "_") {
		prefix = prefix + "_"
	}

	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Filter by prefix if specified
			if e.Prefix != "" {
				if !strings.HasPrefix(key, prefix) {
					continue // Skip env vars that don't match the prefix
				}
				// Remove the prefix and underscore
				key = strings.TrimPrefix(key, prefix)
			}

			// Convert to lowercase to match config tags
			key = strings.ToLower(key)
			result[key] = inferType(value)
		}
	}
	return result, nil
}

// inferType attempts to infer the type of a string value
func inferType(value string) any {
	// Try bool first
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Try int
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Default to string
	return value
}
