package main

// FeatureConfig handles feature flag configuration
// Note: No go:generate directive here - this struct won't have validation generated
type FeatureConfig struct {
	// API features
	RateLimiting bool `yaml:"rate_limiting"`
	Caching      bool `yaml:"caching"`
	Monitoring   bool `yaml:"monitoring"`

	// Security features
	Authentication bool `yaml:"authentication"`
	Authorization  bool `yaml:"authorization"`
	AuditLogging   bool `yaml:"audit_logging"`

	// Experimental features
	ExperimentalFeatures map[string]bool `yaml:"experimental_features"`
}