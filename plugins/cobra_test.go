package plugins

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestLoad_BasicFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("env", "", "environment")
	cmd.Flags().Int("port", 8080, "port number")

	// Simulate CLI flags being set
	_ = cmd.Flags().Set("env", "production")
	_ = cmd.Flags().Set("port", "9090")

	src, err := Cobra(CobraOpts{Cmd: cmd})
	assert.NoError(t, err)
	assert.Equal(t, "cobra", src.Name())

	data, err := src.Load(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"env":  "production",
		"port": "9090", // Cobra stringifies everything
	}, data)
}

func TestLoad_UnsetFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("unset", "", "this won't be changed")

	src, err := Cobra(CobraOpts{Cmd: cmd})
	assert.NoError(t, err)

	data, err := src.Load(context.Background())
	assert.NoError(t, err)
	_, exists := data["unset"]
	assert.False(t, exists, "unset flags shouldn't be loaded")
}

func TestNewSource_InvalidOpts(t *testing.T) {
	_, err := Cobra("not-a-valid-opts")
	assert.Error(t, err)
}
