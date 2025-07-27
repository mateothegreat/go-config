# API Request Validation Example

This example demonstrates using `go-validate` for HTTP API request and response validation, including middleware integration and real-time validation.

## 📋 What You'll Learn

- HTTP request/response validation
- Gin middleware integration
- Real-time API validation
- JSON schema validation
- Error response formatting
- Performance in high-throughput APIs

## 🏗️ Files

- `server.go` - HTTP server with validation middleware
- `handlers.go` - API handlers with request validation
- `middleware.go` - Validation middleware
- `models.go` - Request/response models with validation tags
- `main.go` - Server startup and configuration

## 🚀 Quick Start

1. **Generate validation code:**

   ```bash
   go-validate generate
   ```

2. **Start the server:**

   ```bash
   go run .
   ```

3. **Test API endpoints:**

   ```bash
   # Valid request
   curl -X POST http://localhost:8080/api/users \
     -H "Content-Type: application/json" \
     -d '{"name":"John Doe","email":"john@example.com","age":25}'
   
   # Invalid request (missing email)
   curl -X POST http://localhost:8080/api/users \
     -H "Content-Type: application/json" \
     -d '{"name":"John","age":17}'
   ```

## 📖 API Validation Patterns

### 1. Request Model Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,minlen=2,maxlen=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"min=18,max=120"`
    Password string `json:"password" validate:"required,minlen=8"`
    Role     string `json:"role" validate:"oneof=user|admin|moderator"`
}

func (r *CreateUserRequest) Validate() error {
    // Generated validation method
}
```

### 2. Validation Middleware

```go
func ValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip validation for certain endpoints
        if shouldSkipValidation(c.Request.URL.Path) {
            c.Next()
            return
        }
        
        // Validate request based on content type
        if err := validateRequest(c); err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error:   "Validation failed",
                Details: formatValidationError(err),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 3. Handler Integration

```go
func CreateUserHandler(c *gin.Context) {
    var req CreateUserRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "Invalid JSON format",
        })
        return
    }
    
    // Validate the request
    if err := req.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error:   "Validation failed",
            Details: formatValidationError(err),
        })
        return
    }
    
    // Process valid request
    user, err := createUser(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Failed to create user",
        })
        return
    }
    
    c.JSON(http.StatusCreated, user)
}
```

## 🎯 Advanced API Validation

### 1. Dynamic Validation Rules

```go
type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,minlen=2,maxlen=50"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
    Age   *int    `json:"age,omitempty" validate:"omitempty,min=18,max=120"`
}

func (r *UpdateUserRequest) ValidateForUser(existingUser *User) error {
    // Standard validation first
    if err := r.Validate(); err != nil {
        return err
    }
    
    // Business logic validation
    if r.Email != nil && *r.Email != existingUser.Email {
        if emailExists(*r.Email) {
            return goconfig.NewError().
                Field("Email").
                Format("email already exists")
        }
    }
    
    return nil
}
```

### 2. Conditional Validation

```go
type PaymentRequest struct {
    Amount      float64 `json:"amount" validate:"required,min=0.01"`
    Currency    string  `json:"currency" validate:"required,oneof=USD|EUR|GBP"`
    Method      string  `json:"method" validate:"required,oneof=card|bank|crypto"`
    
    // Card payment fields
    CardNumber string `json:"card_number,omitempty" validate:"required_if=Method:card,creditcard"`
    CardExpiry string `json:"card_expiry,omitempty" validate:"required_if=Method:card"`
    CardCVV    string `json:"card_cvv,omitempty" validate:"required_if=Method:card,len=3"`
    
    // Bank payment fields
    BankAccount string `json:"bank_account,omitempty" validate:"required_if=Method:bank"`
    BankRouting string `json:"bank_routing,omitempty" validate:"required_if=Method:bank"`
    
    // Crypto payment fields
    WalletAddress string `json:"wallet_address,omitempty" validate:"required_if=Method:crypto"`
}
```

### 3. Rate Limiting with Validation

```go
func RateLimitedValidationMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(100, 10) // 100 requests per second, burst of 10
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, ErrorResponse{
                Error: "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        // Perform validation
        if err := validateRequest(c); err != nil {
            // Track validation errors for rate limiting
            trackValidationFailure(c.ClientIP())
            
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error:   "Validation failed",
                Details: formatValidationError(err),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## 🔧 Error Response Formatting

### 1. Structured Error Responses

```go
type ErrorResponse struct {
    Error     string             `json:"error"`
    Details   []ValidationError  `json:"details,omitempty"`
    Code      string             `json:"code"`
    Timestamp time.Time          `json:"timestamp"`
}

type ValidationError struct {
    Field   string      `json:"field"`
    Message string      `json:"message"`
    Value   interface{} `json:"value,omitempty"`
    Code    string      `json:"code"`
}

func formatValidationError(err error) []ValidationError {
    var errors []ValidationError
    
    if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
        for _, verr := range validationErrors.Errors() {
            errors = append(errors, ValidationError{
                Field:   verr.Field,
                Message: verr.Message,
                Value:   verr.Value,
                Code:    verr.Code,
            })
        }
    }
    
    return errors
}
```

### 2. Localized Error Messages

```go
type ErrorLocalizer struct {
    locale   string
    messages map[string]map[string]string
}

func (l *ErrorLocalizer) LocalizeValidationError(err error) []ValidationError {
    errors := formatValidationError(err)
    
    for i, verr := range errors {
        if localized := l.getLocalizedMessage(verr.Code, verr.Field); localized != "" {
            errors[i].Message = localized
        }
    }
    
    return errors
}

func (l *ErrorLocalizer) getLocalizedMessage(code, field string) string {
    if messages, ok := l.messages[l.locale]; ok {
        if template, exists := messages[code]; exists {
            return fmt.Sprintf(template, field)
        }
    }
    return ""
}
```

## 📊 Performance Optimization

### 1. Request Validation Caching

```go
type ValidationCache struct {
    cache map[string]error
    mutex sync.RWMutex
    ttl   time.Duration
}

func (vc *ValidationCache) ValidateWithCache(key string, validator func() error) error {
    vc.mutex.RLock()
    if cached, exists := vc.cache[key]; exists {
        vc.mutex.RUnlock()
        return cached
    }
    vc.mutex.RUnlock()
    
    // Perform validation
    err := validator()
    
    vc.mutex.Lock()
    vc.cache[key] = err
    vc.mutex.Unlock()
    
    // Clean up cache periodically
    go func() {
        time.Sleep(vc.ttl)
        vc.mutex.Lock()
        delete(vc.cache, key)
        vc.mutex.Unlock()
    }()
    
    return err
}
```

### 2. Async Validation

```go
func AsyncValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Quick synchronous validation first
        if err := quickValidation(c); err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error: "Quick validation failed",
            })
            c.Abort()
            return
        }
        
        // Store request for async validation
        go func() {
            if err := deepValidation(c.Copy()); err != nil {
                // Log validation issues for monitoring
                log.Printf("Async validation failed: %v", err)
                // Could trigger alerts or cleanup
            }
        }()
        
        c.Next()
    }
}
```

## 🧪 Testing API Validation

### 1. Validation Test Suite

```go
func TestAPIValidation(t *testing.T) {
    router := setupRouter()
    
    tests := []struct {
        name           string
        method         string
        url            string
        body           interface{}
        expectedStatus int
        expectedErrors []string
    }{
        {
            name:   "valid user creation",
            method: "POST",
            url:    "/api/users",
            body: CreateUserRequest{
                Name:     "John Doe",
                Email:    "john@example.com",
                Age:      25,
                Password: "securepassword",
                Role:     "user",
            },
            expectedStatus: http.StatusCreated,
        },
        {
            name:   "invalid email format",
            method: "POST",
            url:    "/api/users",
            body: CreateUserRequest{
                Name:     "John Doe",
                Email:    "invalid-email",
                Age:      25,
                Password: "securepassword",
                Role:     "user",
            },
            expectedStatus: http.StatusBadRequest,
            expectedErrors: []string{"email"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            body, _ := json.Marshal(tt.body)
            req, _ := http.NewRequest(tt.method, tt.url, bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if len(tt.expectedErrors) > 0 {
                var response ErrorResponse
                err := json.Unmarshal(w.Body.Bytes(), &response)
                require.NoError(t, err)
                
                for _, expectedField := range tt.expectedErrors {
                    found := false
                    for _, detail := range response.Details {
                        if detail.Field == expectedField {
                            found = true
                            break
                        }
                    }
                    assert.True(t, found, "Expected validation error for field: %s", expectedField)
                }
            }
        })
    }
}
```

### 2. Performance Testing

```go
func BenchmarkAPIValidation(b *testing.B) {
    router := setupRouter()
    
    validRequest := CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Age:      25,
        Password: "securepassword",
        Role:     "user",
    }
    
    body, _ := json.Marshal(validRequest)
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        if w.Code != http.StatusCreated {
            b.Fatalf("Expected 201, got %d", w.Code)
        }
    }
}
```

## 🎓 Production Considerations

### 1. Monitoring and Metrics

```go
var (
    validationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "api_validation_duration_seconds",
            Help: "Time spent on request validation",
        },
        []string{"endpoint", "status"},
    )
    
    validationErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_validation_errors_total",
            Help: "Total number of validation errors",
        },
        []string{"endpoint", "field", "error_type"},
    )
)

func instrumentedValidation(endpoint string, validator func() error) error {
    start := time.Now()
    
    err := validator()
    
    status := "success"
    if err != nil {
        status = "error"
        
        // Record specific validation errors
        if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
            for _, verr := range validationErrors.Errors() {
                validationErrors.WithLabelValues(endpoint, verr.Field, verr.Code).Inc()
            }
        }
    }
    
    validationDuration.WithLabelValues(endpoint, status).Observe(time.Since(start).Seconds())
    
    return err
}
```

### 2. Security Considerations

```go
func SecurityValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Rate limiting for validation failures
        if isValidationRateLimited(c.ClientIP()) {
            c.JSON(http.StatusTooManyRequests, ErrorResponse{
                Error: "Too many validation failures",
            })
            c.Abort()
            return
        }
        
        // Input sanitization
        if err := sanitizeRequest(c); err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error: "Invalid input detected",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## 🎓 Best Practices

### 1. API Design

- **Consistent error format** - Standard error response structure
- **Clear field names** - Descriptive JSON field names
- **Appropriate HTTP status codes** - 400 for validation, 422 for business logic
- **Documentation** - Document validation rules in API docs

### 2. Performance

- **Zero-allocation validation** - Generated code avoids allocations
- **Early validation** - Validate before expensive operations
- **Caching** - Cache validation results when appropriate
- **Async validation** - Use async for non-critical validations

### 3. User Experience

- **Helpful error messages** - Explain what needs to be fixed
- **Field-level errors** - Show errors for specific fields
- **Localization** - Support multiple languages
- **Progressive validation** - Validate as user types (client-side)

## 🎓 Next Steps

- Learn error handling patterns in [`../10-error-handling/`](../examples/10-error-handling)
- See multi-package examples in [`../09-multi-package/`](../examples/09-multi-package)
- Explore performance testing in [`../13-performance-testing/`](../examples/13-performance-testing)
