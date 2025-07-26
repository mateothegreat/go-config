package goconfig

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/mateothegreat/go-config/plugins"
)

type ConfigLoader struct {
	mu        sync.RWMutex
	sources   []plugins.Plugin
	target    any
	defaults  any
	validator Validator
	merged    map[string]any
	errors    []error
}

// NewConfigLoader creates a loader for a target struct.
func NewConfigLoader(target any) *ConfigLoader {
	return &ConfigLoader{
		target: target,
	}
}

func (cl *ConfigLoader) Use(src plugins.Plugin) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.sources = append(cl.sources, src)
}

func (cl *ConfigLoader) SetDefaults(def any) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.defaults = def
}

func (cl *ConfigLoader) SetValidator(v Validator) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.validator = v
}

// SetStructValidator sets up struct tag-based validation with optional custom validators
func (cl *ConfigLoader) SetStructValidator(customValidators ...func(*StructValidator)) {
	validator := NewStructValidator()
	for _, fn := range customValidators {
		fn(validator)
	}
	cl.SetValidator(validator)
}

func (cl *ConfigLoader) Load(ctx context.Context) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.merged = make(map[string]any)
	validKeys := getValidConfigKeys(cl.target)

	for _, src := range cl.sources {
		if src == nil {
			cl.errors = append(cl.errors, fmt.Errorf("nil configuration source"))
			continue
		}
		data, err := src.Load(ctx)
		if err != nil {
			cl.errors = append(cl.errors, fmt.Errorf("source %s: %w", src.Name(), err))
			continue
		}
		// Only merge keys that exist in the target struct
		for k, v := range data {
			if validKeys[k] {
				cl.merged[k] = v
			}
		}
	}

	if cl.defaults != nil {
		_ = hydrateIntoMap(cl.defaults, cl.merged, false) // merge defaults only where missing
	}

	if err := hydrateIntoStruct(cl.merged, cl.target); err != nil {
		cl.errors = append(cl.errors, err)
		return err
	}

	if cl.validator != nil {
		if err := cl.validator.Validate(cl.target); err != nil {
			cl.errors = append(cl.errors, err)
			return err
		}
	} else {
		// Use default struct validator if no custom validator is set
		structValidator := NewStructValidator()
		if err := structValidator.Validate(cl.target); err != nil {
			cl.errors = append(cl.errors, err)
			return err
		}
	}

	if len(cl.errors) > 0 {
		return cl.errors[0]
	}

	return nil
}

// Sources returns the names of sources used
func (cl *ConfigLoader) Sources() []string {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	var names []string
	for _, src := range cl.sources {
		names = append(names, src.Name())
	}
	return names
}

// Inspect returns merged keys and errors
func (cl *ConfigLoader) Inspect() map[string]any {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.merged
}

func (cl *ConfigLoader) Errors() []error {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.errors
}

func hydrateIntoMap(from any, into map[string]any, overwrite bool) error {
	val := reflect.ValueOf(from).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		key := field.Tag.Get("config")
		if key == "" {
			key = strings.ToLower(field.Name)
		}
		if _, exists := into[key]; !exists || overwrite {
			into[key] = val.Field(i).Interface()
		}
	}
	return nil
}

func hydrateIntoStruct(from map[string]any, into any) error {
	val := reflect.ValueOf(into)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return fmt.Errorf("target is nil")
	}
	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		key := field.Tag.Get("config")
		if key == "" {
			key = strings.ToLower(field.Name)
		}
		if raw, ok := from[key]; ok {
			dest := val.Field(i)
			if !dest.CanSet() {
				continue
			}

			if err := setFieldValue(dest, raw, field.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func setFieldValue(dest reflect.Value, raw any, fieldName string) error {
	src := reflect.ValueOf(raw)
	destType := dest.Type()

	// Direct assignment if types match
	if src.Type().AssignableTo(destType) {
		dest.Set(src)
		return nil
	}

	// Direct conversion if convertible
	if src.Type().ConvertibleTo(destType) {
		dest.Set(src.Convert(destType))
		return nil
	}

	// Handle string to primitive conversion
	if src.Kind() == reflect.String {
		strVal := src.String()

		switch destType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(strVal, 10, int(destType.Size()*8)); err == nil {
				dest.SetInt(intVal)
				return nil
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if uintVal, err := strconv.ParseUint(strVal, 10, int(destType.Size()*8)); err == nil {
				dest.SetUint(uintVal)
				return nil
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, err := strconv.ParseFloat(strVal, int(destType.Size()*8)); err == nil {
				dest.SetFloat(floatVal)
				return nil
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(strVal); err == nil {
				dest.SetBool(boolVal)
				return nil
			}
		case reflect.String:
			dest.SetString(strVal)
			return nil
		}
	}

	return fmt.Errorf("field %q type mismatch: cannot convert %v (%v) to %v", fieldName, raw, src.Type(), destType)
}

// getValidConfigKeys extracts all valid configuration keys from the target struct
func getValidConfigKeys(target any) map[string]bool {
	validKeys := make(map[string]bool)

	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		key := field.Tag.Get("config")
		if key == "" {
			key = strings.ToLower(field.Name)
		}
		validKeys[key] = true
	}

	return validKeys
}
