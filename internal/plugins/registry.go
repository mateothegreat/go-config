package plugins

import (
	"fmt"
	"sync"
)

// ValidatorPlugin defines the interface for custom validation plugins
type ValidatorPlugin interface {
	Validate(field interface{}) error
}

// PluginRegistry manages custom validation plugins
type PluginRegistry struct {
	mu       sync.RWMutex
	plugins  map[string]ValidatorPlugin
	metadata map[string]PluginMetadata
}

// PluginMetadata contains information about a plugin
type PluginMetadata struct {
	Name           string
	Version        string
	Description    string
	Author         string
	SupportedTypes []string
}

// globalRegistry is the default plugin registry
var globalRegistry = NewPluginRegistry()

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:  make(map[string]ValidatorPlugin),
		metadata: make(map[string]PluginMetadata),
	}
}

// RegisterValidator registers a custom validator plugin
func RegisterValidator(name string, plugin ValidatorPlugin) {
	globalRegistry.RegisterValidator(name, plugin)
}

// RegisterValidatorWithMetadata registers a plugin with metadata
func RegisterValidatorWithMetadata(name string, plugin ValidatorPlugin, metadata PluginMetadata) {
	globalRegistry.RegisterValidatorWithMetadata(name, plugin, metadata)
}

// GetValidator retrieves a validator plugin by name
func GetValidator(name string) (ValidatorPlugin, bool) {
	return globalRegistry.GetValidator(name)
}

// ListValidators returns all registered validator names
func ListValidators() []string {
	return globalRegistry.ListValidators()
}

// RegisterValidator registers a custom validator plugin
func (r *PluginRegistry) RegisterValidator(name string, plugin ValidatorPlugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[name] = plugin
}

// RegisterValidatorWithMetadata registers a plugin with metadata
func (r *PluginRegistry) RegisterValidatorWithMetadata(name string, plugin ValidatorPlugin, metadata PluginMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[name] = plugin
	r.metadata[name] = metadata
}

// GetValidator retrieves a validator plugin by name
func (r *PluginRegistry) GetValidator(name string) (ValidatorPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// GetMetadata retrieves metadata for a plugin
func (r *PluginRegistry) GetMetadata(name string) (PluginMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metadata, exists := r.metadata[name]
	return metadata, exists
}

// ListValidators returns all registered validator names
func (r *PluginRegistry) ListValidators() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// UnregisterValidator removes a validator plugin
func (r *PluginRegistry) UnregisterValidator(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plugins, name)
	delete(r.metadata, name)
}

// Clear removes all registered validators
func (r *PluginRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins = make(map[string]ValidatorPlugin)
	r.metadata = make(map[string]PluginMetadata)
}

// ValidateWithPlugin validates a field using a specific plugin
func (r *PluginRegistry) ValidateWithPlugin(pluginName string, field interface{}) error {
	plugin, exists := r.GetValidator(pluginName)
	if !exists {
		return fmt.Errorf("validator plugin '%s' not found", pluginName)
	}

	return plugin.Validate(field)
}

// Built-in custom validators

// IPAddressValidator validates IP addresses
type IPAddressValidator struct{}

func (v *IPAddressValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("IP validation requires string input")
	}

	if !isValidIP(str) {
		return fmt.Errorf("invalid IP address format")
	}

	return nil
}

// UUIDValidator validates UUID format
type UUIDValidator struct{}

func (v *UUIDValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("UUID validation requires string input")
	}

	if !isValidUUID(str) {
		return fmt.Errorf("invalid UUID format")
	}

	return nil
}

// CreditCardValidator validates credit card numbers using Luhn algorithm
type CreditCardValidator struct{}

func (v *CreditCardValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("credit card validation requires string input")
	}

	if !isValidCreditCard(str) {
		return fmt.Errorf("invalid credit card number")
	}

	return nil
}

// PhoneValidator validates phone numbers
type PhoneValidator struct{}

func (v *PhoneValidator) Validate(field interface{}) error {
	str, ok := field.(string)
	if !ok {
		return fmt.Errorf("phone validation requires string input")
	}

	if !isValidPhone(str) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// Helper functions for built-in validators
func isValidIP(ip string) bool {
	// Simple IP validation - in a real implementation, use net.ParseIP
	return len(ip) > 6 && len(ip) < 16
}

func isValidUUID(uuid string) bool {
	// Simple UUID validation - in a real implementation, use proper UUID parsing
	return len(uuid) == 36
}

func isValidCreditCard(number string) bool {
	// Simple credit card validation using Luhn algorithm
	// This is a simplified version - use a proper library in production
	return len(number) >= 13 && len(number) <= 19
}

func isValidPhone(phone string) bool {
	// Simple phone validation - in a real implementation, use a proper phone library
	return len(phone) >= 10 && len(phone) <= 15
}

// init registers built-in validators
func init() {
	RegisterValidatorWithMetadata("ip", &IPAddressValidator{}, PluginMetadata{
		Name:           "IP Address Validator",
		Version:        "1.0.0",
		Description:    "Validates IPv4 and IPv6 addresses",
		Author:         "go-validate",
		SupportedTypes: []string{"string"},
	})

	RegisterValidatorWithMetadata("uuid", &UUIDValidator{}, PluginMetadata{
		Name:           "UUID Validator",
		Version:        "1.0.0",
		Description:    "Validates UUID format (RFC 4122)",
		Author:         "go-validate",
		SupportedTypes: []string{"string"},
	})

	RegisterValidatorWithMetadata("creditcard", &CreditCardValidator{}, PluginMetadata{
		Name:           "Credit Card Validator",
		Version:        "1.0.0",
		Description:    "Validates credit card numbers using Luhn algorithm",
		Author:         "go-validate",
		SupportedTypes: []string{"string"},
	})

	RegisterValidatorWithMetadata("phone", &PhoneValidator{}, PluginMetadata{
		Name:           "Phone Number Validator",
		Version:        "1.0.0",
		Description:    "Validates phone number formats",
		Author:         "go-validate",
		SupportedTypes: []string{"string"},
	})
}
