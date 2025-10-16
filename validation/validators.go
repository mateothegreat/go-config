package validation

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Pre-compiled regex patterns for maximum performance
var (
	alphaRegex     = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphaNumRegex  = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericRegex   = regexp.MustCompile(`^[0-9]+$`)
	base64Regex    = regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	uuidRegex      = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	hostnameRegex  = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	macRegex       = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
)

// String Validators

// IsRequired checks if a string is not empty
func IsRequired(s string) bool {
	return s != ""
}

// IsEmail validates email addresses using Go's mail package for accuracy
func IsEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsURL validates URLs
func IsURL(u string) bool {
	if u == "" {
		return false
	}
	parsed, err := url.Parse(u)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

// IsAlpha checks if string contains only letters
func IsAlpha(s string) bool {
	return s != "" && alphaRegex.MatchString(s)
}

// IsAlphaNum checks if string contains only letters and numbers
func IsAlphaNum(s string) bool {
	return s != "" && alphaNumRegex.MatchString(s)
}

// IsNumeric checks if string contains only numbers
func IsNumeric(s string) bool {
	return s != "" && numericRegex.MatchString(s)
}

// IsBase64 validates base64 encoding
func IsBase64(s string) bool {
	if s == "" || len(s)%4 != 0 {
		return false
	}
	return base64Regex.MatchString(s)
}

// IsUUID validates UUID format
func IsUUID(s string) bool {
	return s != "" && uuidRegex.MatchString(strings.ToLower(s))
}

// IsIP validates IP addresses (IPv4 or IPv6)
func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

// IsIPv4 validates IPv4 addresses
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

// IsIPv6 validates IPv6 addresses
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}

// IsHostname validates hostnames
func IsHostname(s string) bool {
	return s != "" && len(s) <= 253 && hostnameRegex.MatchString(s)
}

// IsCIDR validates CIDR notation
func IsCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// IsMAC validates MAC addresses
func IsMAC(s string) bool {
	return s != "" && macRegex.MatchString(s)
}

// Length Validators

// HasMinLength checks minimum string length
func HasMinLength(s string, min int) bool {
	return len(s) >= min
}

// HasMaxLength checks maximum string length
func HasMaxLength(s string, max int) bool {
	return len(s) <= max
}

// HasLength checks exact string length
func HasLength(s string, length int) bool {
	return len(s) == length
}

// Numeric Validators

// IsRequiredInt checks if an int is not zero
func IsRequiredInt(i int) bool {
	return i != 0
}

// IsRequiredInt64 checks if an int64 is not zero
func IsRequiredInt64(i int64) bool {
	return i != 0
}

// IsRequiredFloat64 checks if a float64 is not zero
func IsRequiredFloat64(f float64) bool {
	return f != 0.0
}

// IsRequiredFloat32 checks if a float32 is not zero
func IsRequiredFloat32(f float32) bool {
	return f != 0.0
}

// HasMinInt checks minimum int value
func HasMinInt(i, min int) bool {
	return i >= min
}

// HasMaxInt checks maximum int value
func HasMaxInt(i, max int) bool {
	return i <= max
}

// HasMinInt64 checks minimum int64 value
func HasMinInt64(i, min int64) bool {
	return i >= min
}

// HasMaxInt64 checks maximum int64 value
func HasMaxInt64(i, max int64) bool {
	return i <= max
}

// HasMinFloat64 checks minimum float64 value
func HasMinFloat64(f, min float64) bool {
	return f >= min
}

// HasMaxFloat64 checks maximum float64 value
func HasMaxFloat64(f, max float64) bool {
	return f <= max
}

// HasMinFloat32 checks minimum float32 value
func HasMinFloat32(f, min float32) bool {
	return f >= min
}

// HasMaxFloat32 checks maximum float32 value
func HasMaxFloat32(f, max float32) bool {
	return f <= max
}

// Slice Validators

// IsRequiredSlice checks if slice is not empty
func IsRequiredSlice[T any](slice []T) bool {
	return len(slice) > 0
}

// HasMinSliceLength checks minimum slice length
func HasMinSliceLength[T any](slice []T, min int) bool {
	return len(slice) >= min
}

// HasMaxSliceLength checks maximum slice length
func HasMaxSliceLength[T any](slice []T, max int) bool {
	return len(slice) <= max
}

// Bool Validators

// IsRequiredBool checks if bool is true (for required validation)
func IsRequiredBool(b bool) bool {
	return b
}

// OneOf Validators

// IsOneOfString checks if string is one of allowed values
func IsOneOfString(s string, allowed []string) bool {
	for _, a := range allowed {
		if s == a {
			return true
		}
	}
	return false
}

// IsOneOfInt checks if int is one of allowed values
func IsOneOfInt(i int, allowed []int) bool {
	for _, a := range allowed {
		if i == a {
			return true
		}
	}
	return false
}

// Dive Validators (for slice validation)

// ValidateStringSlice validates each string in a slice
func ValidateStringSlice(slice []string, validator func(string) bool) bool {
	for _, s := range slice {
		if !validator(s) {
			return false
		}
	}
	return true
}

// ValidateIntSlice validates each int in a slice
func ValidateIntSlice(slice []int, validator func(int) bool) bool {
	for _, i := range slice {
		if !validator(i) {
			return false
		}
	}
	return true
}

// Complex validation helpers

// ValidateStringWithRules validates a string against multiple rules
func ValidateStringWithRules(s string, rules map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	for rule, param := range rules {
		switch rule {
		case "required":
			if !IsRequired(s) {
				errors |= ErrRequired
			}
		case "email":
			if s != "" && !IsEmail(s) {
				errors |= ErrEmail
			}
		case "url":
			if s != "" && !IsURL(s) {
				errors |= ErrURL
			}
		case "alpha":
			if s != "" && !IsAlpha(s) {
				errors |= ErrAlpha
			}
		case "alphanum":
			if s != "" && !IsAlphaNum(s) {
				errors |= ErrAlphaNum
			}
		case "numeric":
			if s != "" && !IsNumeric(s) {
				errors |= ErrNumeric
			}
		case "base64":
			if s != "" && !IsBase64(s) {
				errors |= ErrBase64
			}
		case "uuid":
			if s != "" && !IsUUID(s) {
				errors |= ErrUUID
			}
		case "ip":
			if s != "" && !IsIP(s) {
				errors |= ErrIP
			}
		case "ipv4":
			if s != "" && !IsIPv4(s) {
				errors |= ErrIPv4
			}
		case "ipv6":
			if s != "" && !IsIPv6(s) {
				errors |= ErrIPv6
			}
		case "hostname":
			if s != "" && !IsHostname(s) {
				errors |= ErrHostname
			}
		case "cidr":
			if s != "" && !IsCIDR(s) {
				errors |= ErrCIDR
			}
		case "mac":
			if s != "" && !IsMAC(s) {
				errors |= ErrMAC
			}
		case "minlen":
			if min, ok := param.(int); ok && !HasMinLength(s, min) {
				errors |= ErrMinLength
			}
		case "maxlen":
			if max, ok := param.(int); ok && !HasMaxLength(s, max) {
				errors |= ErrMaxLength
			}
		case "len":
			if length, ok := param.(int); ok && !HasLength(s, length) {
				errors |= ErrLength
			}
		case "oneof":
			if allowed, ok := param.([]string); ok && !IsOneOfString(s, allowed) {
				errors |= ErrOneOf
			}
		}
	}

	return errors
}

// ValidateIntWithRules validates an int against multiple rules
func ValidateIntWithRules(i int, rules map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	for rule, param := range rules {
		switch rule {
		case "required":
			if !IsRequiredInt(i) {
				errors |= ErrRequired
			}
		case "min":
			if min, ok := param.(int); ok && !HasMinInt(i, min) {
				errors |= ErrMin
			}
		case "max":
			if max, ok := param.(int); ok && !HasMaxInt(i, max) {
				errors |= ErrMax
			}
		case "oneof":
			if allowed, ok := param.([]int); ok && !IsOneOfInt(i, allowed) {
				errors |= ErrOneOf
			}
		}
	}

	return errors
}

// Helper functions for string parsing

// ParseInt safely parses string to int
func ParseInt(s string) (int, bool) {
	i, err := strconv.Atoi(s)
	return i, err == nil
}

// ParseFloat64 safely parses string to float64
func ParseFloat64(s string) (float64, bool) {
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

// ParseBool safely parses string to bool
func ParseBool(s string) (bool, bool) {
	b, err := strconv.ParseBool(s)
	return b, err == nil
}

// IsASCII checks if string contains only ASCII characters
func IsASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// IsLowercase checks if string is all lowercase
func IsLowercase(s string) bool {
	return s == strings.ToLower(s)
}

// IsUppercase checks if string is all uppercase
func IsUppercase(s string) bool {
	return s == strings.ToUpper(s)
}