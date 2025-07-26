package goconfig

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidationErrors represents a collection of validation errors.
type ValidationErrors []string

// Error implements the error interface.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(ve, "; "))
}

// Errors returns the slice of error strings for iteration.
func (ve ValidationErrors) Errors() []string {
	return []string(ve)
}

// HasErrors returns true if there are any validation errors.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// AsValidationErrors attempts to convert an error to ValidationErrors.
// Returns the ValidationErrors and true if successful, nil and false otherwise.
func AsValidationErrors(err error) (ValidationErrors, bool) {
	if validationErr, ok := err.(ValidationErrors); ok {
		return validationErr, true
	}
	return nil, false
}

type Validator interface {
	Validate(data any) error
}

// FieldValidator represents a single validation rule for a field.
type FieldValidator func(fieldName string, value reflect.Value, rule string) error

// TypedValidator represents type-specific validation functions that avoid reflection.
type TypedValidator interface {
	ValidateString(fieldName string, value string, rules map[string]string) []string
	ValidateInt(fieldName string, value int, rules map[string]string) []string
	ValidateBool(fieldName string, value bool, rules map[string]string) []string
	ValidateFloat(fieldName string, value float64, rules map[string]string) []string
}

// FastValidator implements TypedValidator with minimal reflection.
type FastValidator struct {
	stringValidators map[string]func(string, string, string) error
	intValidators    map[string]func(string, int, string) error
	boolValidators   map[string]func(string, bool, string) error
	floatValidators  map[string]func(string, float64, string) error
}

// NewFastValidator creates a new fast validator with type-specific functions.
func NewFastValidator() *FastValidator {
	fv := &FastValidator{
		stringValidators: make(map[string]func(string, string, string) error),
		intValidators:    make(map[string]func(string, int, string) error),
		boolValidators:   make(map[string]func(string, bool, string) error),
		floatValidators:  make(map[string]func(string, float64, string) error),
	}

	// Register string validators
	fv.stringValidators["required"] = validateStringRequired
	fv.stringValidators["minlen"] = validateStringMinLen
	fv.stringValidators["maxlen"] = validateStringMaxLen
	fv.stringValidators["len"] = validateStringLen
	fv.stringValidators["regex"] = validateStringRegex
	fv.stringValidators["oneof"] = validateStringOneOf
	fv.stringValidators["email"] = validateStringEmail
	fv.stringValidators["url"] = validateStringURL
	fv.stringValidators["alpha"] = validateStringAlpha
	fv.stringValidators["alphanumeric"] = validateStringAlphaNumeric
	fv.stringValidators["numeric"] = validateStringNumeric

	// Register int validators
	fv.intValidators["required"] = validateIntRequired
	fv.intValidators["min"] = validateIntMin
	fv.intValidators["max"] = validateIntMax
	fv.intValidators["range"] = validateIntRange

	// Register float validators
	fv.floatValidators["required"] = validateFloatRequired
	fv.floatValidators["min"] = validateFloatMin
	fv.floatValidators["max"] = validateFloatMax
	fv.floatValidators["range"] = validateFloatRange

	return fv
}

// ValidateString validates a string value using registered string validators
func (fv *FastValidator) ValidateString(fieldName string, value string, rules map[string]string) []string {
	var errors []string
	for ruleName, ruleValue := range rules {
		if validator, exists := fv.stringValidators[ruleName]; exists {
			if err := validator(fieldName, value, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}
	return errors
}

// ValidateInt validates an int value using registered int validators
func (fv *FastValidator) ValidateInt(fieldName string, value int, rules map[string]string) []string {
	var errors []string
	for ruleName, ruleValue := range rules {
		if validator, exists := fv.intValidators[ruleName]; exists {
			if err := validator(fieldName, value, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}
	return errors
}

// ValidateBool validates a bool value (minimal validation needed)
func (fv *FastValidator) ValidateBool(fieldName string, value bool, rules map[string]string) []string {
	var errors []string
	// Bool validation is typically limited, but we can check for required
	if _, required := rules["required"]; required && !value {
		errors = append(errors, fmt.Sprintf("field '%s' is required", fieldName))
	}
	return errors
}

// ValidateFloat validates a float64 value using registered float validators
func (fv *FastValidator) ValidateFloat(fieldName string, value float64, rules map[string]string) []string {
	var errors []string
	for ruleName, ruleValue := range rules {
		if validator, exists := fv.floatValidators[ruleName]; exists {
			if err := validator(fieldName, value, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}
	return errors
}

// ValidationRegistry holds all available validators.
type ValidationRegistry struct {
	validators map[string]FieldValidator
}

// NewValidationRegistry creates a new registry with built-in validators.
func NewValidationRegistry() *ValidationRegistry {
	registry := &ValidationRegistry{
		validators: make(map[string]FieldValidator),
	}

	// Register built-in validators.
	registry.RegisterValidator("required", validateRequired)
	registry.RegisterValidator("min", validateMin)
	registry.RegisterValidator("max", validateMax)
	registry.RegisterValidator("minlen", validateMinLen)
	registry.RegisterValidator("maxlen", validateMaxLen)
	registry.RegisterValidator("len", validateLen)
	registry.RegisterValidator("regex", validateRegex)
	registry.RegisterValidator("oneof", validateOneOf)
	registry.RegisterValidator("range", validateRange)
	registry.RegisterValidator("email", validateEmail)
	registry.RegisterValidator("url", validateURL)
	registry.RegisterValidator("alpha", validateAlpha)
	registry.RegisterValidator("alphanumeric", validateAlphaNumeric)
	registry.RegisterValidator("numeric", validateNumeric)

	return registry
}

// RegisterValidator adds a custom validator to the registry.
func (r *ValidationRegistry) RegisterValidator(name string, validator FieldValidator) {
	r.validators[name] = validator
}

// StructValidator implements the Validator interface using struct tags.
type StructValidator struct {
	registry *ValidationRegistry
}

// NewStructValidator creates a new struct validator with default registry.
func NewStructValidator() *StructValidator {
	return &StructValidator{
		registry: NewValidationRegistry(),
	}
}

// WithCustomValidator adds a custom validator function.
func (v *StructValidator) WithCustomValidator(name string, validator FieldValidator) *StructValidator {
	v.registry.RegisterValidator(name, validator)
	return v
}

// FastStructValidator implements validation with minimal reflection
type FastStructValidator struct {
	fastValidator *FastValidator
}

// NewFastStructValidator creates a new fast struct validator
func NewFastStructValidator() *FastStructValidator {
	return &FastStructValidator{
		fastValidator: NewFastValidator(),
	}
}

// Validate validates a struct using minimal reflection and type-specific validators
func (fsv *FastStructValidator) Validate(data any) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validation target must be a struct, got %v", val.Kind())
	}

	typ := val.Type()
	var errors ValidationErrors

	// Use cached field info to reduce reflection overhead
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		rules := parseValidationRules(validateTag)
		fieldErrors := fsv.validateFieldByType(field.Name, fieldValue, rules)
		errors = append(errors, fieldErrors...)
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateFieldByType validates a field based on its type using minimal reflection
func (fsv *FastStructValidator) validateFieldByType(fieldName string, fieldValue reflect.Value, rules map[string]string) []string {
	// Switch on the field's kind to minimize reflection calls
	switch fieldValue.Kind() {
	case reflect.String:
		return fsv.fastValidator.ValidateString(fieldName, fieldValue.String(), rules)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fsv.fastValidator.ValidateInt(fieldName, int(fieldValue.Int()), rules)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fsv.fastValidator.ValidateInt(fieldName, int(fieldValue.Uint()), rules)
	case reflect.Float32, reflect.Float64:
		return fsv.fastValidator.ValidateFloat(fieldName, fieldValue.Float(), rules)
	case reflect.Bool:
		return fsv.fastValidator.ValidateBool(fieldName, fieldValue.Bool(), rules)
	default:
		// For unsupported types, fall back to string representation
		return fsv.fastValidator.ValidateString(fieldName, fmt.Sprintf("%v", fieldValue.Interface()), rules)
	}
}

// Validate validates a struct using validation tags.
func (v *StructValidator) Validate(data any) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validation target must be a struct, got %v", val.Kind())
	}

	typ := val.Type()
	var errors ValidationErrors

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		rules := parseValidationRules(validateTag)
		for ruleName, ruleValue := range rules {
			validator, exists := v.registry.validators[ruleName]
			if !exists {
				errors = append(errors, fmt.Sprintf("unknown validator '%s' for field '%s'", ruleName, field.Name))
				continue
			}

			if err := validator(field.Name, fieldValue, ruleValue); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// parseValidationRules parses validation tag string into rules map.
func parseValidationRules(tag string) map[string]string {
	rules := make(map[string]string)
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			rules[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else {
			rules[part] = ""
		}
	}

	return rules
}

// Built-in validators.

func validateRequired(fieldName string, value reflect.Value, _ string) error {
	if isZero(value) {
		return fmt.Errorf("field '%s' is required", fieldName)
	}
	return nil
}

func validateMin(fieldName string, value reflect.Value, rule string) error {
	min, err := strconv.ParseFloat(rule, 64)
	if err != nil {
		return fmt.Errorf("invalid min rule '%s' for field '%s'", rule, fieldName)
	}

	val, err := getNumericValue(value)
	if err != nil {
		return fmt.Errorf("field '%s' must be numeric for min validation", fieldName)
	}

	if val < min {
		return fmt.Errorf("field '%s' must be at least %g", fieldName, min)
	}
	return nil
}

func validateMax(fieldName string, value reflect.Value, rule string) error {
	max, err := strconv.ParseFloat(rule, 64)
	if err != nil {
		return fmt.Errorf("invalid max rule '%s' for field '%s'", rule, fieldName)
	}

	val, err := getNumericValue(value)
	if err != nil {
		return fmt.Errorf("field '%s' must be numeric for max validation", fieldName)
	}

	if val > max {
		return fmt.Errorf("field '%s' must be at most %g", fieldName, max)
	}
	return nil
}

func validateMinLen(fieldName string, value reflect.Value, rule string) error {
	minLen, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid minlen rule '%s' for field '%s'", rule, fieldName)
	}

	length := getLength(value)
	if length < minLen {
		return fmt.Errorf("field '%s' must be at least %d characters long", fieldName, minLen)
	}
	return nil
}

func validateMaxLen(fieldName string, value reflect.Value, rule string) error {
	maxLen, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid maxlen rule '%s' for field '%s'", rule, fieldName)
	}

	length := getLength(value)
	if length > maxLen {
		return fmt.Errorf("field '%s' must be at most %d characters long", fieldName, maxLen)
	}
	return nil
}

func validateLen(fieldName string, value reflect.Value, rule string) error {
	expectedLen, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid len rule '%s' for field '%s'", rule, fieldName)
	}

	length := getLength(value)
	if length != expectedLen {
		return fmt.Errorf("field '%s' must be exactly %d characters long", fieldName, expectedLen)
	}
	return nil
}

func validateRegex(fieldName string, value reflect.Value, rule string) error {
	regex, err := regexp.Compile(rule)
	if err != nil {
		return fmt.Errorf("invalid regex rule '%s' for field '%s': %v", rule, fieldName, err)
	}

	str := getString(value)
	if !regex.MatchString(str) {
		return fmt.Errorf("field '%s' does not match pattern '%s'", fieldName, rule)
	}
	return nil
}

func validateOneOf(fieldName string, value reflect.Value, rule string) error {
	options := strings.Split(rule, "|")
	str := getString(value)

	for _, option := range options {
		if strings.TrimSpace(option) == str {
			return nil
		}
	}

	return fmt.Errorf("field '%s' must be one of [%s]", fieldName, rule)
}

func validateRange(fieldName string, value reflect.Value, rule string) error {
	parts := strings.Split(rule, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range rule '%s' for field '%s', expected 'min:max'", rule, fieldName)
	}

	min, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid range min '%s' for field '%s'", parts[0], fieldName)
	}

	max, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid range max '%s' for field '%s'", parts[1], fieldName)
	}

	val, err := getNumericValue(value)
	if err != nil {
		return fmt.Errorf("field '%s' must be numeric for range validation", fieldName)
	}

	if val < min || val > max {
		return fmt.Errorf("field '%s' must be between %g and %g", fieldName, min, max)
	}
	return nil
}

func validateEmail(fieldName string, value reflect.Value, rule string) error {
	str := getString(value)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("field '%s' must be a valid email address", fieldName)
	}
	return nil
}

func validateURL(fieldName string, value reflect.Value, rule string) error {
	str := getString(value)
	// More flexible URL regex that accepts various schemes
	urlRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^\s]*$`)
	if !urlRegex.MatchString(str) {
		return fmt.Errorf("field '%s' must be a valid URL", fieldName)
	}
	return nil
}

func validateAlpha(fieldName string, value reflect.Value, rule string) error {
	str := getString(value)
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(str) {
		return fmt.Errorf("field '%s' must contain only alphabetic characters", fieldName)
	}
	return nil
}

func validateAlphaNumeric(fieldName string, value reflect.Value, rule string) error {
	str := getString(value)
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphaNumRegex.MatchString(str) {
		return fmt.Errorf("field '%s' must contain only alphanumeric characters", fieldName)
	}
	return nil
}

func validateNumeric(fieldName string, value reflect.Value, rule string) error {
	str := getString(value)
	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if !numericRegex.MatchString(str) {
		return fmt.Errorf("field '%s' must contain only numeric characters", fieldName)
	}
	return nil
}

// Helper functions.

func isZero(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Slice, reflect.Map, reflect.Array:
		return value.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return value.IsNil()
	}
	return value.Interface() == reflect.Zero(value.Type()).Interface()
}

func getNumericValue(value reflect.Value) (float64, error) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	default:
		return 0, fmt.Errorf("not a numeric type")
	}
}

func getLength(value reflect.Value) int {
	switch value.Kind() {
	case reflect.String:
		return len(value.String())
	case reflect.Slice, reflect.Array, reflect.Map:
		return value.Len()
	default:
		return len(fmt.Sprintf("%v", value.Interface()))
	}
}

func getString(value reflect.Value) string {
	if value.Kind() == reflect.String {
		return value.String()
	}
	return fmt.Sprintf("%v", value.Interface())
}

// Type-specific validation functions (no reflection)

// String validators
func validateStringRequired(fieldName string, value string, _ string) error {
	if value == "" {
		return fmt.Errorf("field '%s' is required", fieldName)
	}
	return nil
}

func validateStringMinLen(fieldName string, value string, rule string) error {
	minLen, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid minlen rule '%s' for field '%s'", rule, fieldName)
	}
	if len(value) < minLen {
		return fmt.Errorf("field '%s' must be at least %d characters long", fieldName, minLen)
	}
	return nil
}

func validateStringMaxLen(fieldName string, value string, rule string) error {
	maxLen, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid maxlen rule '%s' for field '%s'", rule, fieldName)
	}
	if len(value) > maxLen {
		return fmt.Errorf("field '%s' must be at most %d characters long", fieldName, maxLen)
	}
	return nil
}

func validateStringLen(fieldName string, value string, rule string) error {
	expectedLen, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid len rule '%s' for field '%s'", rule, fieldName)
	}
	if len(value) != expectedLen {
		return fmt.Errorf("field '%s' must be exactly %d characters long", fieldName, expectedLen)
	}
	return nil
}

func validateStringRegex(fieldName string, value string, rule string) error {
	regex, err := regexp.Compile(rule)
	if err != nil {
		return fmt.Errorf("invalid regex rule '%s' for field '%s': %v", rule, fieldName, err)
	}
	if !regex.MatchString(value) {
		return fmt.Errorf("field '%s' does not match pattern '%s'", fieldName, rule)
	}
	return nil
}

func validateStringOneOf(fieldName string, value string, rule string) error {
	options := strings.Split(rule, "|")
	for _, option := range options {
		if strings.TrimSpace(option) == value {
			return nil
		}
	}
	return fmt.Errorf("field '%s' must be one of [%s]", fieldName, rule)
}

func validateStringEmail(fieldName string, value string, _ string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		return fmt.Errorf("field '%s' must be a valid email address", fieldName)
	}
	return nil
}

func validateStringURL(fieldName string, value string, _ string) error {
	urlRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^\s]*$`)
	if !urlRegex.MatchString(value) {
		return fmt.Errorf("field '%s' must be a valid URL", fieldName)
	}
	return nil
}

func validateStringAlpha(fieldName string, value string, _ string) error {
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(value) {
		return fmt.Errorf("field '%s' must contain only alphabetic characters", fieldName)
	}
	return nil
}

func validateStringAlphaNumeric(fieldName string, value string, _ string) error {
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphaNumRegex.MatchString(value) {
		return fmt.Errorf("field '%s' must contain only alphanumeric characters", fieldName)
	}
	return nil
}

func validateStringNumeric(fieldName string, value string, _ string) error {
	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if !numericRegex.MatchString(value) {
		return fmt.Errorf("field '%s' must contain only numeric characters", fieldName)
	}
	return nil
}

// Int validators
func validateIntRequired(fieldName string, value int, _ string) error {
	if value == 0 {
		return fmt.Errorf("field '%s' is required", fieldName)
	}
	return nil
}

func validateIntMin(fieldName string, value int, rule string) error {
	min, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid min rule '%s' for field '%s'", rule, fieldName)
	}
	if value < min {
		return fmt.Errorf("field '%s' must be at least %d", fieldName, min)
	}
	return nil
}

func validateIntMax(fieldName string, value int, rule string) error {
	max, err := strconv.Atoi(rule)
	if err != nil {
		return fmt.Errorf("invalid max rule '%s' for field '%s'", rule, fieldName)
	}
	if value > max {
		return fmt.Errorf("field '%s' must be at most %d", fieldName, max)
	}
	return nil
}

func validateIntRange(fieldName string, value int, rule string) error {
	parts := strings.Split(rule, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range rule '%s' for field '%s', expected 'min:max'", rule, fieldName)
	}
	
	min, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid range min '%s' for field '%s'", parts[0], fieldName)
	}
	
	max, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid range max '%s' for field '%s'", parts[1], fieldName)
	}
	
	if value < min || value > max {
		return fmt.Errorf("field '%s' must be between %d and %d", fieldName, min, max)
	}
	return nil
}

// Float validators
func validateFloatRequired(fieldName string, value float64, _ string) error {
	if value == 0 {
		return fmt.Errorf("field '%s' is required", fieldName)
	}
	return nil
}

func validateFloatMin(fieldName string, value float64, rule string) error {
	min, err := strconv.ParseFloat(rule, 64)
	if err != nil {
		return fmt.Errorf("invalid min rule '%s' for field '%s'", rule, fieldName)
	}
	if value < min {
		return fmt.Errorf("field '%s' must be at least %g", fieldName, min)
	}
	return nil
}

func validateFloatMax(fieldName string, value float64, rule string) error {
	max, err := strconv.ParseFloat(rule, 64)
	if err != nil {
		return fmt.Errorf("invalid max rule '%s' for field '%s'", rule, fieldName)
	}
	if value > max {
		return fmt.Errorf("field '%s' must be at most %g", fieldName, max)
	}
	return nil
}

func validateFloatRange(fieldName string, value float64, rule string) error {
	parts := strings.Split(rule, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range rule '%s' for field '%s', expected 'min:max'", rule, fieldName)
	}
	
	min, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid range min '%s' for field '%s'", parts[0], fieldName)
	}
	
	max, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid range max '%s' for field '%s'", parts[1], fieldName)
	}
	
	if value < min || value > max {
		return fmt.Errorf("field '%s' must be between %g and %g", fieldName, min, max)
	}
	return nil
}
