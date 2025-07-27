// Package validators provides built-in validation plugins
package validators

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// IPAddressValidator validates IP addresses
type IPAddressValidator struct{}

func (v *IPAddressValidator) Name() string { return "ip" }

func (v *IPAddressValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("IP validation requires string input")
	}

	if net.ParseIP(str) == nil {
		return fmt.Errorf("invalid IP address format")
	}

	return nil
}

func (v *IPAddressValidator) SupportedTypes() []string {
	return []string{"string"}
}

// Local interface definitions to avoid import cycles
type ValidatorPlugin interface {
	Name() string
	Validate(field interface{}) error
	SupportedTypes() []string
}

type PluginMetadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Type        string   `json:"type"`
	Features    []string `json:"features,omitempty"`
}

type ValidatorFactory interface {
	Create() ValidatorPlugin
	Metadata() PluginMetadata
}

// UUIDValidator validates UUID format
type UUIDValidator struct{}

func (v *UUIDValidator) Name() string { return "uuid" }

func (v *UUIDValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("UUID validation requires string input")
	}

	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(str) {
		return fmt.Errorf("invalid UUID format")
	}

	return nil
}

func (v *UUIDValidator) SupportedTypes() []string {
	return []string{"string"}
}

// CreditCardValidator validates credit card numbers using Luhn algorithm
type CreditCardValidator struct{}

func (v *CreditCardValidator) Name() string { return "creditcard" }

func (v *CreditCardValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("credit card validation requires string input")
	}

	// Remove spaces and dashes
	clean := strings.ReplaceAll(strings.ReplaceAll(str, " ", ""), "-", "")

	if !isValidCreditCard(clean) {
		return fmt.Errorf("invalid credit card number")
	}

	return nil
}

func (v *CreditCardValidator) SupportedTypes() []string {
	return []string{"string"}
}

// PhoneValidator validates phone numbers
type PhoneValidator struct{}

func (v *PhoneValidator) Name() string { return "phone" }

func (v *PhoneValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("phone validation requires string input")
	}

	// Basic phone validation - remove non-digits and check length
	digits := regexp.MustCompile(`\D`).ReplaceAllString(str, "")
	if len(digits) < 10 || len(digits) > 15 {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

func (v *PhoneValidator) SupportedTypes() []string {
	return []string{"string"}
}

// Validator factories

type IPValidatorFactory struct{}

func (f *IPValidatorFactory) Create() ValidatorPlugin { return &IPAddressValidator{} }
func (f *IPValidatorFactory) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        "ip",
		Version:     "1.0.0",
		Description: "Validates IPv4 and IPv6 addresses",
		Author:      "go-config",
		Type:        "validator",
		Features:    []string{"network", "format"},
	}
}

type UUIDValidatorFactory struct{}

func (f *UUIDValidatorFactory) Create() ValidatorPlugin { return &UUIDValidator{} }
func (f *UUIDValidatorFactory) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        "uuid",
		Version:     "1.0.0",
		Description: "Validates UUID format (RFC 4122)",
		Author:      "go-config",
		Type:        "validator",
		Features:    []string{"format", "identifier"},
	}
}

type CreditCardValidatorFactory struct{}

func (f *CreditCardValidatorFactory) Create() ValidatorPlugin { return &CreditCardValidator{} }
func (f *CreditCardValidatorFactory) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        "creditcard",
		Version:     "1.0.0",
		Description: "Validates credit card numbers using Luhn algorithm",
		Author:      "go-config",
		Type:        "validator",
		Features:    []string{"financial", "luhn"},
	}
}

type PhoneValidatorFactory struct{}

func (f *PhoneValidatorFactory) Create() ValidatorPlugin { return &PhoneValidator{} }
func (f *PhoneValidatorFactory) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        "phone",
		Version:     "1.0.0",
		Description: "Validates phone number formats",
		Author:      "go-config",
		Type:        "validator",
		Features:    []string{"telecommunication", "format"},
	}
}

// Helper functions

func isValidCreditCard(number string) bool {
	// Luhn algorithm implementation
	if len(number) < 13 || len(number) > 19 {
		return false
	}

	sum := 0
	alternate := false

	// Process digits from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
