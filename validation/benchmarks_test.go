package validation

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// BenchmarkValidationErrors_BitField benchmarks bit-field error operations
func BenchmarkValidationErrors_BitField(b *testing.B) {
	b.Run("Add", func(b *testing.B) {
		var errors ValidationErrors
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errors.Add(ErrRequired)
		}
	})

	b.Run("HasError", func(b *testing.B) {
		var errors ValidationErrors
		errors.Add(ErrRequired)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = errors.HasError(ErrRequired)
		}
	})

	b.Run("Count", func(b *testing.B) {
		var errors ValidationErrors
		errors |= ErrRequired | ErrEmail | ErrURL
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = errors.Count()
		}
	})
}

// BenchmarkValidationErrors_SliceComparison benchmarks slice-based error tracking for comparison
func BenchmarkValidationErrors_SliceComparison(b *testing.B) {
	type SliceErrors []string

	hasError := func(errors SliceErrors, err string) bool {
		for _, e := range errors {
			if e == err {
				return true
			}
		}
		return false
	}

	addError := func(errors *SliceErrors, err string) {
		if !hasError(*errors, err) {
			*errors = append(*errors, err)
		}
	}

	b.Run("Add", func(b *testing.B) {
		var errors SliceErrors
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			addError(&errors, "required")
		}
	})

	b.Run("HasError", func(b *testing.B) {
		var errors SliceErrors
		addError(&errors, "required")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = hasError(errors, "required")
		}
	})

	b.Run("Count", func(b *testing.B) {
		var errors SliceErrors
		addError(&errors, "required")
		addError(&errors, "email")
		addError(&errors, "url")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = len(errors)
		}
	})
}

// BenchmarkStringValidators benchmarks string validation functions
func BenchmarkStringValidators(b *testing.B) {
	testCases := map[string]struct {
		input     string
		validator func(string) bool
	}{
		"IsRequired":  {"test", IsRequired},
		"IsEmail":     {"test@example.com", IsEmail},
		"IsURL":       {"https://example.com", IsURL},
		"IsAlpha":     {"abcXYZ", IsAlpha},
		"IsAlphaNum":  {"abc123XYZ", IsAlphaNum},
		"IsNumeric":   {"12345", IsNumeric},
		"IsBase64":    {"SGVsbG8gV29ybGQ=", IsBase64},
		"IsUUID":      {"550e8400-e29b-41d4-a716-446655440000", IsUUID},
		"IsIP":        {"192.168.1.1", IsIP},
		"IsIPv4":      {"192.168.1.1", IsIPv4},
		"IsIPv6":      {"2001:0db8:85a3::8a2e:370:7334", IsIPv6},
		"IsHostname":  {"example.com", IsHostname},
		"IsCIDR":      {"192.168.1.0/24", IsCIDR},
		"IsMAC":       {"00:11:22:33:44:55", IsMAC},
	}

	for name, tc := range testCases {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tc.validator(tc.input)
			}
		})
	}
}

// BenchmarkReflectionVsDirectValidation compares reflection-based vs direct validation
func BenchmarkReflectionVsDirectValidation(b *testing.B) {
	type TestStruct struct {
		Name  string `validate:"required,minlen=3"`
		Email string `validate:"required,email"`
		Age   int    `validate:"required,min=18"`
	}

	testData := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	// Direct validation (optimized)
	b.Run("Direct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var errors ValidationErrors
			if !IsRequired(testData.Name) {
				errors |= ErrRequired
			}
			if !HasMinLength(testData.Name, 3) {
				errors |= ErrMinLength
			}
			if !IsRequired(testData.Email) {
				errors |= ErrRequired
			}
			if testData.Email != "" && !IsEmail(testData.Email) {
				errors |= ErrEmail
			}
			if !IsRequiredInt(testData.Age) {
				errors |= ErrRequired
			}
			if !HasMinInt(testData.Age, 18) {
				errors |= ErrMin
			}
			_ = errors
		}
	})

	// Reflection-based validation (for comparison)
	b.Run("Reflection", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validateWithReflection(testData)
		}
	})
}

// validateWithReflection simulates reflection-based validation
func validateWithReflection(obj interface{}) []string {
	var errors []string
	val := reflect.ValueOf(obj)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		tag := field.Tag.Get("validate")

		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			if strings.Contains(rule, "=") {
				parts := strings.SplitN(rule, "=", 2)
				ruleName := parts[0]
				ruleValue := parts[1]

				switch ruleName {
				case "minlen":
					if value.Kind() == reflect.String {
						minLen, _ := strconv.Atoi(ruleValue)
						if len(value.String()) < minLen {
							errors = append(errors, "minlen validation failed")
						}
					}
				case "min":
					if value.Kind() == reflect.Int {
						min, _ := strconv.Atoi(ruleValue)
						if int(value.Int()) < min {
							errors = append(errors, "min validation failed")
						}
					}
				}
			} else {
				switch rule {
				case "required":
					switch value.Kind() {
					case reflect.String:
						if value.String() == "" {
							errors = append(errors, "required validation failed")
						}
					case reflect.Int:
						if value.Int() == 0 {
							errors = append(errors, "required validation failed")
						}
					}
				case "email":
					if value.Kind() == reflect.String {
						// Simplified email check for benchmark
						email := value.String()
						if email != "" && !strings.Contains(email, "@") {
							errors = append(errors, "email validation failed")
						}
					}
				}
			}
		}
	}

	return errors
}

// BenchmarkComplexValidation benchmarks complex validation scenarios
func BenchmarkComplexValidation(b *testing.B) {
	rules := map[string]interface{}{
		"required": true,
		"minlen":   3,
		"maxlen":   50,
		"email":    true,
	}

	validEmail := "test@example.com"
	invalidEmail := "invalid-email"

	b.Run("ValidStringWithRules", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ValidateStringWithRules(validEmail, rules)
		}
	})

	b.Run("InvalidStringWithRules", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ValidateStringWithRules(invalidEmail, rules)
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("ValidationErrors", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var errors ValidationErrors
			errors |= ErrRequired
			errors |= ErrEmail
			errors |= ErrURL
			_ = errors.HasErrors()
			_ = errors.Count()
		}
	})

	b.Run("SliceErrors", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var errors []string
			errors = append(errors, "required")
			errors = append(errors, "email")
			errors = append(errors, "url")
			_ = len(errors) > 0
			_ = len(errors)
		}
	})

	b.Run("ValidationResult", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := NewValidationResult()
			result.AddFieldError("name", ErrRequired)
			result.AddFieldError("email", ErrEmail)
			_ = result.HasErrors()
		}
	})
}

// BenchmarkLengthValidators benchmarks length validation functions
func BenchmarkLengthValidators(b *testing.B) {
	testString := "Hello, World! This is a test string for benchmarking."

	b.Run("HasMinLength", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMinLength(testString, 10)
		}
	})

	b.Run("HasMaxLength", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMaxLength(testString, 100)
		}
	})

	b.Run("HasLength", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasLength(testString, len(testString))
		}
	})
}

// BenchmarkNumericValidators benchmarks numeric validation functions
func BenchmarkNumericValidators(b *testing.B) {
	testInt := 42
	testFloat := 42.5

	b.Run("IsRequiredInt", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = IsRequiredInt(testInt)
		}
	})

	b.Run("HasMinInt", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMinInt(testInt, 10)
		}
	})

	b.Run("HasMaxInt", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMaxInt(testInt, 100)
		}
	})

	b.Run("HasMinFloat64", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMinFloat64(testFloat, 10.0)
		}
	})

	b.Run("HasMaxFloat64", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMaxFloat64(testFloat, 100.0)
		}
	})
}

// BenchmarkSliceValidators benchmarks slice validation functions
func BenchmarkSliceValidators(b *testing.B) {
	testSlice := []string{"apple", "banana", "cherry", "date", "elderberry"}

	b.Run("IsRequiredSlice", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = IsRequiredSlice(testSlice)
		}
	})

	b.Run("HasMinSliceLength", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMinSliceLength(testSlice, 3)
		}
	})

	b.Run("HasMaxSliceLength", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = HasMaxSliceLength(testSlice, 10)
		}
	})

	b.Run("ValidateStringSlice", func(b *testing.B) {
		validator := func(s string) bool { return len(s) > 3 }
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ValidateStringSlice(testSlice, validator)
		}
	})
}