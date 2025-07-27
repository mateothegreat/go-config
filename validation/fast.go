package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FastValidator implements TypedValidator with minimal reflection
type FastValidator struct {
	stringValidators map[string]func(string, string, string) error
	intValidators    map[string]func(string, int, string) error
	boolValidators   map[string]func(string, bool, string) error
	floatValidators  map[string]func(string, float64, string) error
}

// NewFastValidator creates a new fast validator with type-specific functions
func NewFastValidator() TypedValidator {
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
