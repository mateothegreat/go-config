package examples

type SimpleConfig struct {
	Port int    `config:"port" validate:"required,min=1000,max=65535"`
	Env  string `config:"env" validate:"required,oneof=development|staging|production"`
}

type AdvancedConfig struct {
	// Server configuration
	Host     string `config:"host" validate:"required,regex=^[a-zA-Z0-9.-]+$"`
	Port     int    `config:"port" validate:"required,range=1000:65535"`
	Protocol string `config:"protocol" validate:"oneof=http|https"`
	
	// Database configuration  
	DatabaseURL string `config:"database_url" validate:"required,url"`
	MaxConns    int    `config:"max_conns" validate:"min=1,max=100"`
	
	// Authentication
	JWTSecret   string `config:"jwt_secret" validate:"required,minlen=32"`
	AdminEmail  string `config:"admin_email" validate:"required,email"`
	
	// Features
	EnableDebug   bool   `config:"enable_debug"`
	LogLevel      string `config:"log_level" validate:"oneof=debug|info|warn|error"`
	Version       string `config:"version" validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
	
	// API Keys
	APIKey        string `config:"api_key" validate:"required,len=40,alphanumeric"`
	WebhookSecret string `config:"webhook_secret" validate:"required,minlen=16,maxlen=64"`
}
