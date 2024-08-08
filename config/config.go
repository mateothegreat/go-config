package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/risersh/util/files"
	"github.com/risersh/util/validation"
)

type GetConfigArgs struct {
	Paths     []string
	WalkDepth int
}

// GetConfig returns a config of type T.
// It will merge the base config with the environment config.
// If the environment config does not exist, it will use the base config.
//
// Arguments:
//   - GetConfigArgs: The arguments to use.
//
// Returns:
//   - A pointer to the config of type T.
//   - An error if the config could not be found.
func GetConfig[T any](args GetConfigArgs) (*T, error) {
	config := new(T)

	// Read and merge all configs from the provided paths.
	for _, path := range args.Paths {
		tempConfig := new(T)
		configPath := files.WalkFile(path, args.WalkDepth)
		if configPath != "" {
			// Attempt to read the config from the file.
			if err := cleanenv.ReadConfig(configPath, tempConfig); err != nil {
				return nil, fmt.Errorf("failed to read config from %s: %w", configPath, err)
			}
		} else {
			// Read the environment variables into the tempConfig.
			if err := cleanenv.ReadEnv(tempConfig); err != nil {
				return nil, fmt.Errorf("failed to read environment variables: %w", err)
			}
		}

		// Merge the tempConfig into the config.
		MergeStructs(config, tempConfig)
	}

	// Validate the final config struct.
	emptyFields, err := validation.ValidateStructFields(config, "")
	if err != nil {
		return nil, err
	}
	if len(emptyFields) > 0 {
		return nil, fmt.Errorf("empty fields: %v", emptyFields)
	}

	return config, nil
}

// SaveConfig saves the config to the path.
// It will walk the path to find the config file and save it to the file.
//
// Arguments:
//   - config interface{}: The config to save.
//   - path string: The path to save the config to.
//   - walkDepth int: The depth to walk the path to find the config file.
func SaveConfig(config interface{}, path string, walkDepth int) error {
	configPath := files.WalkFile(path, walkDepth)
	if configPath != "" {
		str, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return os.WriteFile(configPath, str, 0644)
	}
	return fmt.Errorf("config path not found")
}

// MergeStructs recursively merges src into dst.
//
// Arguments:
//   - dst: The destination struct to merge the source into.
//   - src: The source struct to merge into the destination.
func MergeStructs(dst, src interface{}) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldName := srcVal.Type().Field(i).Name

		if dstField := dstVal.FieldByName(srcFieldName); dstField.IsValid() && dstField.CanSet() {
			if srcField.Kind() == reflect.Struct {
				MergeStructs(dstField.Addr().Interface(), srcField.Addr().Interface())
			} else if dstField.IsZero() {
				dstField.Set(srcField)
			}
		}
	}
}
