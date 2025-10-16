package validation

import (
	"testing"
)

func TestValidationErrors(t *testing.T) {
	t.Run("HasErrors", func(t *testing.T) {
		var errors ValidationErrors
		if errors.HasErrors() {
			t.Error("Expected no errors, but HasErrors returned true")
		}

		errors |= ErrRequired
		if !errors.HasErrors() {
			t.Error("Expected errors, but HasErrors returned false")
		}
	})

	t.Run("HasError", func(t *testing.T) {
		var errors ValidationErrors
		errors |= ErrRequired | ErrEmail

		if !errors.HasError(ErrRequired) {
			t.Error("Expected ErrRequired to be present")
		}

		if !errors.HasError(ErrEmail) {
			t.Error("Expected ErrEmail to be present")
		}

		if errors.HasError(ErrURL) {
			t.Error("Expected ErrURL to not be present")
		}
	})

	t.Run("Add", func(t *testing.T) {
		var errors ValidationErrors
		errors.Add(ErrRequired)
		errors.Add(ErrEmail)

		if !errors.HasError(ErrRequired) {
			t.Error("Expected ErrRequired to be present after Add")
		}

		if !errors.HasError(ErrEmail) {
			t.Error("Expected ErrEmail to be present after Add")
		}
	})

	t.Run("Remove", func(t *testing.T) {
		var errors ValidationErrors
		errors |= ErrRequired | ErrEmail

		errors.Remove(ErrRequired)
		if errors.HasError(ErrRequired) {
			t.Error("Expected ErrRequired to be removed")
		}

		if !errors.HasError(ErrEmail) {
			t.Error("Expected ErrEmail to still be present")
		}
	})

	t.Run("Count", func(t *testing.T) {
		var errors ValidationErrors
		if errors.Count() != 0 {
			t.Errorf("Expected count 0, got %d", errors.Count())
		}

		errors |= ErrRequired
		if errors.Count() != 1 {
			t.Errorf("Expected count 1, got %d", errors.Count())
		}

		errors |= ErrEmail
		if errors.Count() != 2 {
			t.Errorf("Expected count 2, got %d", errors.Count())
		}
	})

	t.Run("Error", func(t *testing.T) {
		var errors ValidationErrors
		if errors.Error() != "" {
			t.Error("Expected empty error message for no errors")
		}

		errors |= ErrRequired
		if errors.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("ErrorList", func(t *testing.T) {
		var errors ValidationErrors
		errors |= ErrRequired | ErrEmail

		errorList := errors.ErrorList()
		if len(errorList) != 2 {
			t.Errorf("Expected 2 errors in list, got %d", len(errorList))
		}
	})
}

func TestStringValidators(t *testing.T) {
	tests := []struct {
		name      string
		validator func(string) bool
		input     string
		expected  bool
	}{
		{"IsRequired empty", IsRequired, "", false},
		{"IsRequired non-empty", IsRequired, "test", true},
		{"IsEmail valid", IsEmail, "test@example.com", true},
		{"IsEmail invalid", IsEmail, "invalid-email", false},
		{"IsEmail empty", IsEmail, "", false},
		{"IsURL valid", IsURL, "https://example.com", true},
		{"IsURL invalid", IsURL, "not-a-url", false},
		{"IsURL empty", IsURL, "", false},
		{"IsAlpha valid", IsAlpha, "abcXYZ", true},
		{"IsAlpha invalid", IsAlpha, "abc123", false},
		{"IsAlpha empty", IsAlpha, "", false},
		{"IsAlphaNum valid", IsAlphaNum, "abc123XYZ", true},
		{"IsAlphaNum invalid", IsAlphaNum, "abc-123", false},
		{"IsNumeric valid", IsNumeric, "12345", true},
		{"IsNumeric invalid", IsNumeric, "12a45", false},
		{"IsBase64 valid", IsBase64, "SGVsbG8gV29ybGQ=", true},
		{"IsBase64 invalid", IsBase64, "invalid!", false},
		{"IsUUID valid", IsUUID, "550e8400-e29b-41d4-a716-446655440000", true},
		{"IsUUID invalid", IsUUID, "not-a-uuid", false},
		{"IsIP valid IPv4", IsIP, "192.168.1.1", true},
		{"IsIP valid IPv6", IsIP, "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"IsIP invalid", IsIP, "not-an-ip", false},
		{"IsIPv4 valid", IsIPv4, "192.168.1.1", true},
		{"IsIPv4 invalid", IsIPv4, "2001:0db8:85a3::8a2e:370:7334", false},
		{"IsIPv6 valid", IsIPv6, "2001:0db8:85a3::8a2e:370:7334", true},
		{"IsIPv6 invalid", IsIPv6, "192.168.1.1", false},
		{"IsHostname valid", IsHostname, "example.com", true},
		{"IsHostname invalid", IsHostname, "invalid_hostname", false},
		{"IsCIDR valid", IsCIDR, "192.168.1.0/24", true},
		{"IsCIDR invalid", IsCIDR, "not-cidr", false},
		{"IsMAC valid", IsMAC, "00:11:22:33:44:55", true},
		{"IsMAC invalid", IsMAC, "invalid-mac", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.validator(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input %q", tt.expected, result, tt.input)
			}
		})
	}
}

func TestLengthValidators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		param    int
		expected bool
	}{
		{"HasMinLength valid", "hello", 3, true},
		{"HasMinLength invalid", "hi", 3, false},
		{"HasMaxLength valid", "hello", 10, true},
		{"HasMaxLength invalid", "hello world", 5, false},
		{"HasLength valid", "hello", 5, true},
		{"HasLength invalid", "hello", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			switch tt.name {
			case "HasMinLength valid", "HasMinLength invalid":
				result = HasMinLength(tt.input, tt.param)
			case "HasMaxLength valid", "HasMaxLength invalid":
				result = HasMaxLength(tt.input, tt.param)
			case "HasLength valid", "HasLength invalid":
				result = HasLength(tt.input, tt.param)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input %q with param %d", tt.expected, result, tt.input, tt.param)
			}
		})
	}
}

func TestNumericValidators(t *testing.T) {
	t.Run("IsRequiredInt", func(t *testing.T) {
		if IsRequiredInt(0) {
			t.Error("Expected false for zero value")
		}
		if !IsRequiredInt(42) {
			t.Error("Expected true for non-zero value")
		}
	})

	t.Run("HasMinInt", func(t *testing.T) {
		if !HasMinInt(10, 5) {
			t.Error("Expected true for value >= min")
		}
		if HasMinInt(3, 5) {
			t.Error("Expected false for value < min")
		}
	})

	t.Run("HasMaxInt", func(t *testing.T) {
		if !HasMaxInt(5, 10) {
			t.Error("Expected true for value <= max")
		}
		if HasMaxInt(15, 10) {
			t.Error("Expected false for value > max")
		}
	})
}

func TestSliceValidators(t *testing.T) {
	t.Run("IsRequiredSlice", func(t *testing.T) {
		if IsRequiredSlice([]string{}) {
			t.Error("Expected false for empty slice")
		}
		if !IsRequiredSlice([]string{"test"}) {
			t.Error("Expected true for non-empty slice")
		}
	})

	t.Run("HasMinSliceLength", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		if !HasMinSliceLength(slice, 2) {
			t.Error("Expected true for slice length >= min")
		}
		if HasMinSliceLength(slice, 5) {
			t.Error("Expected false for slice length < min")
		}
	})

	t.Run("HasMaxSliceLength", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		if !HasMaxSliceLength(slice, 5) {
			t.Error("Expected true for slice length <= max")
		}
		if HasMaxSliceLength(slice, 2) {
			t.Error("Expected false for slice length > max")
		}
	})
}

func TestOneOfValidators(t *testing.T) {
	t.Run("IsOneOfString", func(t *testing.T) {
		allowed := []string{"red", "green", "blue"}
		if !IsOneOfString("red", allowed) {
			t.Error("Expected true for allowed value")
		}
		if IsOneOfString("yellow", allowed) {
			t.Error("Expected false for non-allowed value")
		}
	})

	t.Run("IsOneOfInt", func(t *testing.T) {
		allowed := []int{1, 2, 3}
		if !IsOneOfInt(2, allowed) {
			t.Error("Expected true for allowed value")
		}
		if IsOneOfInt(5, allowed) {
			t.Error("Expected false for non-allowed value")
		}
	})
}

func TestValidationResult(t *testing.T) {
	t.Run("NewValidationResult", func(t *testing.T) {
		result := NewValidationResult()
		if result == nil {
			t.Error("Expected non-nil result")
		}
		if result.HasErrors() {
			t.Error("Expected no errors in new result")
		}
	})

	t.Run("AddFieldError", func(t *testing.T) {
		result := NewValidationResult()
		result.AddFieldError("name", ErrRequired)

		if !result.HasErrors() {
			t.Error("Expected errors after adding field error")
		}

		fieldErrors := result.GetFieldErrors("name")
		if !fieldErrors.HasError(ErrRequired) {
			t.Error("Expected ErrRequired for field 'name'")
		}
	})

	t.Run("Error", func(t *testing.T) {
		result := NewValidationResult()
		if result.Error() != "" {
			t.Error("Expected empty error message for no errors")
		}

		result.AddFieldError("name", ErrRequired)
		if result.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	})
}

func TestComplexValidation(t *testing.T) {
	t.Run("ValidateStringWithRules", func(t *testing.T) {
		rules := map[string]interface{}{
			"required": true,
			"minlen":   3,
			"email":    true,
		}

		// Valid email
		errors := ValidateStringWithRules("test@example.com", rules)
		if errors.HasErrors() {
			t.Errorf("Expected no errors for valid email, got: %v", errors)
		}

		// Invalid email
		errors = ValidateStringWithRules("invalid", rules)
		if !errors.HasError(ErrEmail) {
			t.Error("Expected email error for invalid email")
		}

		// Empty string (required)
		errors = ValidateStringWithRules("", rules)
		if !errors.HasError(ErrRequired) {
			t.Error("Expected required error for empty string")
		}

		// Too short
		errors = ValidateStringWithRules("x@", rules)
		if !errors.HasError(ErrMinLength) {
			t.Error("Expected min length error for short string")
		}
	})

	t.Run("ValidateIntWithRules", func(t *testing.T) {
		rules := map[string]interface{}{
			"required": true,
			"min":      10,
			"max":      100,
		}

		// Valid int
		errors := ValidateIntWithRules(50, rules)
		if errors.HasErrors() {
			t.Errorf("Expected no errors for valid int, got: %v", errors)
		}

		// Zero (required)
		errors = ValidateIntWithRules(0, rules)
		if !errors.HasError(ErrRequired) {
			t.Error("Expected required error for zero value")
		}

		// Too small
		errors = ValidateIntWithRules(5, rules)
		if !errors.HasError(ErrMin) {
			t.Error("Expected min error for small value")
		}

		// Too large
		errors = ValidateIntWithRules(150, rules)
		if !errors.HasError(ErrMax) {
			t.Error("Expected max error for large value")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("ParseInt", func(t *testing.T) {
		i, ok := ParseInt("123")
		if !ok || i != 123 {
			t.Errorf("Expected (123, true), got (%d, %v)", i, ok)
		}

		_, ok = ParseInt("not-a-number")
		if ok {
			t.Error("Expected false for invalid int")
		}
	})

	t.Run("ParseFloat64", func(t *testing.T) {
		f, ok := ParseFloat64("123.45")
		if !ok || f != 123.45 {
			t.Errorf("Expected (123.45, true), got (%f, %v)", f, ok)
		}

		_, ok = ParseFloat64("not-a-number")
		if ok {
			t.Error("Expected false for invalid float")
		}
	})

	t.Run("ParseBool", func(t *testing.T) {
		b, ok := ParseBool("true")
		if !ok || !b {
			t.Errorf("Expected (true, true), got (%v, %v)", b, ok)
		}

		_, ok = ParseBool("not-a-bool")
		if ok {
			t.Error("Expected false for invalid bool")
		}
	})

	t.Run("IsASCII", func(t *testing.T) {
		if !IsASCII("Hello World") {
			t.Error("Expected true for ASCII string")
		}
		if IsASCII("Hello 世界") {
			t.Error("Expected false for non-ASCII string")
		}
	})

	t.Run("IsLowercase", func(t *testing.T) {
		if !IsLowercase("hello") {
			t.Error("Expected true for lowercase string")
		}
		if IsLowercase("Hello") {
			t.Error("Expected false for mixed case string")
		}
	})

	t.Run("IsUppercase", func(t *testing.T) {
		if !IsUppercase("HELLO") {
			t.Error("Expected true for uppercase string")
		}
		if IsUppercase("Hello") {
			t.Error("Expected false for mixed case string")
		}
	})
}