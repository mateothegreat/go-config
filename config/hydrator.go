package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// structHydrator implements Hydrator using reflection
type structHydrator struct {
	strategy      HydrationStrategy
	caseSensitive bool
}

// NewHydrator creates a new hydrator with the specified strategy
func NewHydrator(strategy HydrationStrategy) Hydrator {
	return &structHydrator{
		strategy:      strategy,
		caseSensitive: false,
	}
}

// Hydrate converts raw configuration data into the target struct
func (h *structHydrator) Hydrate(data map[string]any, target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct, got %v", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		key := h.getFieldKey(field)
		if raw, exists := data[key]; exists {
			if err := h.setFieldValue(fieldValue, raw, field.Name); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// GetValidKeys returns the valid configuration keys for a target struct
func (h *structHydrator) GetValidKeys(target any) map[string]bool {
	validKeys := make(map[string]bool)

	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return validKeys
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		key := h.getFieldKey(field)
		validKeys[key] = true
	}

	return validKeys
}

// getFieldKey determines the configuration key for a struct field
func (h *structHydrator) getFieldKey(field reflect.StructField) string {
	switch h.strategy {
	case HydrationTags:
		// Check for config tag first, then yaml, then json
		if configTag := field.Tag.Get("config"); configTag != "" {
			return configTag
		}
		if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
			// Parse out just the field name, ignoring options like ",omitempty"
			if parts := strings.Split(yamlTag, ","); len(parts) > 0 {
				return parts[0]
			}
			return yamlTag
		}
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Parse out just the field name, ignoring options like ",omitempty"
			if parts := strings.Split(jsonTag, ","); len(parts) > 0 {
				return parts[0]
			}
			return jsonTag
		}
		fallthrough
	case HydrationReflection:
		// Use field name, converted to lowercase if not case sensitive
		name := field.Name
		if !h.caseSensitive {
			name = strings.ToLower(name)
		}
		return name
	case HydrationAuto:
		// For auto, prefer tags if available, otherwise use field name
		if yamlTag := field.Tag.Get("yaml"); yamlTag != "" && yamlTag != "-" {
			// Parse out just the field name, ignoring options like ",omitempty"
			if parts := strings.Split(yamlTag, ","); len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			// Parse out just the field name, ignoring options like ",omitempty"
			if parts := strings.Split(jsonTag, ","); len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
		// Fallback to lowercase field name
		name := field.Name
		if !h.caseSensitive {
			name = strings.ToLower(name)
		}
		return name
	default:
		return strings.ToLower(field.Name)
	}
}

// setFieldValue sets a struct field value from raw configuration data
func (h *structHydrator) setFieldValue(dest reflect.Value, raw any, fieldName string) error {
	src := reflect.ValueOf(raw)
	destType := dest.Type()

	// Handle nil values
	if !src.IsValid() || (src.Kind() == reflect.Ptr && src.IsNil()) {
		return nil
	}

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
		return h.convertStringToType(dest, src.String(), fieldName)
	}

	// Handle nested struct conversion
	if dest.Kind() == reflect.Struct && src.Kind() == reflect.Map {
		return h.hydrateNestedStruct(dest, raw, fieldName)
	}

	// Handle slice conversion
	if dest.Kind() == reflect.Slice && src.Kind() == reflect.Slice {
		return h.convertSlice(dest, src, fieldName)
	}

	return fmt.Errorf("field %q type mismatch: cannot convert %v (%v) to %v", fieldName, raw, src.Type(), destType)
}

// convertStringToType converts string values to the target type
func (h *structHydrator) convertStringToType(dest reflect.Value, strVal, fieldName string) error {
	destType := dest.Type()

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

	return fmt.Errorf("cannot convert string %q to %v", strVal, destType)
}

// hydrateNestedStruct handles nested struct hydration
func (h *structHydrator) hydrateNestedStruct(dest reflect.Value, raw any, fieldName string) error {
	dataMap, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map for nested struct %s, got %T", fieldName, raw)
	}

	// Create a new instance if needed
	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		return h.Hydrate(dataMap, dest.Interface())
	}

	// For non-pointer structs, we need to create a pointer temporarily
	ptr := dest.Addr()
	return h.Hydrate(dataMap, ptr.Interface())
}

// convertSlice handles slice conversion
func (h *structHydrator) convertSlice(dest, src reflect.Value, fieldName string) error {
	srcLen := src.Len()
	destType := dest.Type()

	// Create new slice with appropriate capacity
	newSlice := reflect.MakeSlice(destType, srcLen, srcLen)

	for i := 0; i < srcLen; i++ {
		srcElem := src.Index(i)
		destElem := newSlice.Index(i)

		if err := h.setFieldValue(destElem, srcElem.Interface(), fmt.Sprintf("%s[%d]", fieldName, i)); err != nil {
			return err
		}
	}

	dest.Set(newSlice)
	return nil
}

// structToMap converts a struct to a map (helper function)
func structToMap(data any) (map[string]any, error) {
	result := make(map[string]any)

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		// Get the key using similar logic to getFieldKey
		key := field.Tag.Get("config")
		if key == "" {
			key = field.Tag.Get("yaml")
		}
		if key == "" {
			key = field.Tag.Get("json")
		}
		if key == "" {
			key = strings.ToLower(field.Name)
		}

		result[key] = fieldValue.Interface()
	}

	return result, nil
}
