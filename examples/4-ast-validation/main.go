package main

import (
	"fmt"

	"github.com/mateothegreat/go-validation"
)

func main() {
	// Test the generated validation directly
	fmt.Println("Testing AST-generated validation code...")

	// Test 1: Valid configuration
	// validConfig := &ServerConfig{
	// 	Name:         "api-server",
	// 	Environment:  "prod",
	// 	Host:         "0.0.0.0",
	// 	Port:         8080,
	// 	PublicURL:    "https://api.example.com",
	// 	APIKey:       "YXNkZmFzZGZhc2RmYXNkZmFzZGZhc2RmYXNkZmFzZGY=",
	// 	SecretKey:    "1234567890123456789012345678901234567890123456789012345678901234",
	// 	AllowedCIDRs: []string{"10.0.0.0/8", "172.16.0.0/12"},
	// 	Database: DatabaseConfig{
	// 		Driver:       "postgres",
	// 		Host:         "db.example.com",
	// 		Port:         5432,
	// 		Username:     "dbuser",
	// 		Password:     "dbpass",
	// 		Database:     "myapp",
	// 		MaxOpenConns: 25,
	// 		MaxIdleConns: 10,
	// 	},
	// 	MaxConnections: 1000,
	// 	Timeout:        30.5,
	// }
	validConfig := &EmailConfig{
		FromAddress:  "test@example.com",
		ReplyTo:      "reply@example.com",
		AdminEmails:  []string{"admin@example.com"},
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "username",
		SMTPPassword: "password",
		UseTLS:       true,
	}

	if err := validConfig.Validate(); err != nil {
		fmt.Printf("❌ Valid config failed validation: %v\n", err)
	} else {
		fmt.Println("✅ Valid configuration passed validation!")
	}

	// Test 2: Invalid configuration
	fmt.Println("\nTesting invalid configuration...")
	invalidConfig := &EmailConfig{
		FromAddress:  "test@example.com",
		ReplyTo:      "reply@example.com",
		AdminEmails:  []string{"admin@examEQW@#$%ple.com"},
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "username",
		SMTPPassword: "password",
		UseTLS:       true,
	}

	if err := invalidConfig.Validate(); err != nil {
		fmt.Printf("✅ Invalid config failed validation (as expected):\n%v\n", err)
	} else {
		fmt.Println("❌ Invalid configuration should have failed validation!")
	}

	// Test 3: Test with go-validation library
	fmt.Println("\nTesting with go-validation library...")
	if err := validation.New().Struct(validConfig); err != nil {
		fmt.Printf("❌ go-validation failed: %v\n", err)
	} else {
		fmt.Println("✅ go-validation passed!")
	}
}
