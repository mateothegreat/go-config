package main

import (
	"strings"
	"testing"

	"github.com/mateothegreat/go-validation"
)

func TestServerConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *ServerConfig
		wantError bool
		wantErrs  []string
	}{
		{
			name: "valid configuration",
			config: &ServerConfig{
				Name:         "api-server",
				Environment:  "prod",
				Host:         "192.168.1.1",
				Port:         8080,
				PublicURL:    "https://api.example.com",
				APIKey:       "YXNkZmFzZGZhc2RmYXNkZmFzZGZhc2RmYXNkZmFzZGY=",
				SecretKey:    "1234567890123456789012345678901234567890123456789012345678901234",
				AllowedCIDRs: []string{"10.0.0.0/8", "172.16.0.0/12"},
				Database: DatabaseConfig{
					Driver:       "postgres",
					Host:         "db.example.com",
					Port:         5432,
					Username:     "dbuser",
					Password:     "dbpass",
					Database:     "myapp",
					MaxOpenConns: 25,
					MaxIdleConns: 10,
				},
				MaxConnections: 1000,
				Timeout:        30.5,
			},
			wantError: false,
		},
		{
			name: "name validation errors",
			config: &ServerConfig{
				Name: "ab", // Too short
			},
			wantError: true,
			wantErrs:  []string{"Name", "must be at least 3 characters"},
		},
		{
			name: "environment validation",
			config: &ServerConfig{
				Name:        "valid-name",
				Environment: "development", // Not in allowed values
			},
			wantError: true,
			wantErrs:  []string{"Environment", "must be one of"},
		},
		{
			name: "port range validation",
			config: &ServerConfig{
				Name:        "valid-name",
				Environment: "dev",
				Port:        70000, // Too high
			},
			wantError: true,
			wantErrs:  []string{"Port", "must be at most 65535"},
		},
		{
			name: "email validation",
			config: &ServerConfig{
				Name:        "valid-name",
				Environment: "dev",
				Port:        8080,
			},
			wantError: true,
			wantErrs:  []string{"Host", "is required"},
		},
		{
			name: "CIDR validation",
			config: &ServerConfig{
				Name:         "valid-name",
				Environment:  "dev",
				Host:         "localhost",
				Port:         8080,
				PublicURL:    "https://api.example.com",
				APIKey:       "YXNkZmFzZGZhc2RmYXNkZmFzZGZhc2RmYXNkZmFzZGY=",
				SecretKey:    "1234567890123456789012345678901234567890123456789012345678901234",
				AllowedCIDRs: []string{"invalid-cidr"},
			},
			wantError: true,
			wantErrs:  []string{"AllowedCIDRs", "must be a valid cidr"},
		},
		{
			name: "multiple validation errors",
			config: &ServerConfig{
				Name:        "x",       // Too short
				Environment: "invalid", // Not in oneof
				Port:        70000,     // Too high
			},
			wantError: true,
			wantErrs: []string{
				"Name", "must be at least 3",
				"Environment", "must be one of",
				"Port", "must be at most 65535",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Struct(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected validation error but got none")
					return
				}

				errStr := err.Error()
				for _, wantErr := range tt.wantErrs {
					if !strings.Contains(errStr, wantErr) {
						t.Errorf("expected error to contain %q, got: %s", wantErr, errStr)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestDatabaseConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *DatabaseConfig
		wantError bool
		wantErrs  []string
	}{
		{
			name: "valid postgres config",
			config: &DatabaseConfig{
				Driver:       "postgres",
				Host:         "db.example.com",
				Port:         5432,
				Username:     "user",
				Password:     "pass",
				Database:     "mydb",
				MaxOpenConns: 25,
				MaxIdleConns: 10,
			},
			wantError: false,
		},
		{
			name: "valid sqlite config",
			config: &DatabaseConfig{
				Driver:   "sqlite",
				Database: "data.db",
				// Even though these shouldn't be required for sqlite, without required_if they are validated
				Host:         "localhost",
				Port:         1,
				MaxOpenConns: 1,
			},
			wantError: false,
		},
		{
			name: "invalid driver",
			config: &DatabaseConfig{
				Driver:   "oracle", // Not in allowed values
				Database: "test",   // Required field
			},
			wantError: true,
			wantErrs:  []string{"Driver", "must be one of"},
		},
		{
			name: "missing database field",
			config: &DatabaseConfig{
				Driver: "sqlite",
				// Missing database field which is always required
			},
			wantError: true,
			wantErrs:  []string{"Database", "is required"},
		},
		{
			name: "connection pool validation",
			config: &DatabaseConfig{
				Driver:       "postgres",
				Host:         "db.example.com",
				Port:         5432,
				Username:     "user",
				Password:     "pass",
				Database:     "mydb",
				MaxOpenConns: 0,   // Too low
				MaxIdleConns: 101, // Too high
			},
			wantError: true,
			wantErrs:  []string{"MaxOpenConns", "must be at least 1", "MaxIdleConns", "must be at most 100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test DatabaseConfig validation directly
			err := validation.Struct(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected validation error but got none")
					return
				}

				errStr := err.Error()
				for _, wantErr := range tt.wantErrs {
					if !strings.Contains(errStr, wantErr) {
						t.Errorf("expected error to contain %q, got: %s", wantErr, errStr)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestEmailConfigValidation(t *testing.T) {
	validEmails := []string{"admin@example.com", "support@example.com"}

	tests := []struct {
		name      string
		config    *EmailConfig
		wantError bool
		wantErrs  []string
	}{
		{
			name: "valid email config",
			config: &EmailConfig{
				FromAddress:  "noreply@example.com",
				ReplyTo:      "support@example.com",
				AdminEmails:  validEmails,
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUsername: "smtp-user",
				SMTPPassword: "smtp-pass",
				UseTLS:       true,
			},
			wantError: false,
		},
		{
			name: "invalid email addresses",
			config: &EmailConfig{
				FromAddress: "not-an-email",
				ReplyTo:     "also-not-an-email",
				AdminEmails: []string{"invalid-email", "another-invalid"},
			},
			wantError: true,
			wantErrs:  []string{"FromAddress", "must be a valid email", "ReplyTo", "AdminEmails"},
		},
		{
			name: "empty admin emails",
			config: &EmailConfig{
				FromAddress:  "noreply@example.com",
				AdminEmails:  []string{},
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUsername: "user",
				SMTPPassword: "pass",
			},
			wantError: true,
			wantErrs:  []string{"AdminEmails", "must have at least 1 items"},
		},
		{
			name: "invalid SMTP port",
			config: &EmailConfig{
				FromAddress:  "noreply@example.com",
				AdminEmails:  validEmails,
				SMTPHost:     "smtp.example.com",
				SMTPPort:     100000, // Too high
				SMTPUsername: "user",
				SMTPPassword: "pass",
			},
			wantError: true,
			wantErrs:  []string{"SMTPPort", "must be at most 65535"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError {
				if err == nil {
					t.Errorf("expected validation error but got none")
					return
				}

				errStr := err.Error()
				for _, wantErr := range tt.wantErrs {
					if !strings.Contains(errStr, wantErr) {
						t.Errorf("expected error to contain %q, got: %s", wantErr, errStr)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func BenchmarkValidation(b *testing.B) {
	config := &ServerConfig{
		Name:         "api-server",
		Environment:  "prod",
		Host:         "192.168.1.1",
		Port:         8080,
		PublicURL:    "https://api.example.com",
		APIKey:       "YXNkZmFzZGZhc2RmYXNkZmFzZGZhc2RmYXNkZmFzZGY=",
		SecretKey:    "1234567890123456789012345678901234567890123456789012345678901234",
		AllowedCIDRs: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		Database: DatabaseConfig{
			Driver:       "postgres",
			Host:         "db.example.com",
			Port:         5432,
			Username:     "dbuser",
			Password:     "dbpass",
			Database:     "myapp",
			MaxOpenConns: 25,
			MaxIdleConns: 10,
		},
		MaxConnections: 1000,
		Timeout:        30.5,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validation.Struct(config)
	}
}

func BenchmarkValidationWithErrors(b *testing.B) {
	config := &ServerConfig{
		Name:        "x",       // Too short
		Environment: "invalid", // Not in allowed
		Port:        70000,     // Too high
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validation.Struct(config)
	}
}
