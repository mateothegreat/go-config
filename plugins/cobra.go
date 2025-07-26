package plugins

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CobraOpts struct {
	Cmd *cobra.Command
}

type CobraSource struct {
	cmd *cobra.Command
}

func Cobra(opts any) (Plugin, error) {
	typed, ok := opts.(CobraOpts)
	if !ok {
		return nil, errors.New("invalid Cobra options")
	}
	return &CobraSource{cmd: typed.Cmd}, nil
}

func (cs *CobraSource) Name() string {
	return "cobra"
}

func (cs *CobraSource) Load(ctx context.Context) (map[string]any, error) {
	result := make(map[string]any)
	flags := cs.cmd.Flags()

	flags.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			result[f.Name] = f.Value.String()
		}
	})

	return result, nil
}
