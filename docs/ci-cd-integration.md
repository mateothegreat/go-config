# CI/CD Integration Example - Production Deployment

This example demonstrates comprehensive CI/CD pipeline integration for `go-validate` in production environments, including GitHub Actions, Docker, and automated deployment strategies.

## 📋 What You'll Learn

- GitHub Actions workflow integration
- Docker multi-stage builds with validation
- Automated code generation verification
- Production deployment strategies
- Security scanning and compliance
- Performance testing in CI/CD

## 🏗️ Files

- `.github/workflows/` - GitHub Actions workflows
- `docker/` - Docker configurations
- `scripts/` - Build and deployment scripts
- `configs/` - Environment-specific configurations
- `Makefile` - Build automation
- `docker-compose.yml` - Development environment

## 🚀 Quick Start

1. **Fork/clone the repository**
2. **Set up GitHub secrets** (see Security section)
3. **Push changes** to trigger workflows
4. **Deploy with Docker:**

   ```bash
   docker-compose up --build
   ```

## 📖 GitHub Actions Workflows

### 1. Comprehensive Validation Workflow

```yaml
# .github/workflows/validate.yml
name: Validate and Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'

jobs:
  validate-generation:
    name: Validate Code Generation
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
        
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
          
    - name: Install go-validate
      run: go install github.com/mateothegreat/go-config/cmd/go-validate@latest
      
    - name: Generate validation code
      run: |
        go generate ./...
        go-validate generate --verbose
        
    - name: Verify generated code is current
      run: |
        git add .
        if ! git diff --cached --quiet; then
          echo "❌ Generated code is out of date"
          echo "Files changed:"
          git diff --cached --name-only
          echo "Please run 'go generate ./...' and commit the changes"
          exit 1
        fi
        echo "✅ Generated code is up to date"
        
    - name: Verify code compiles
      run: go build ./...
      
    - name: Run tests
      run: go test -race -coverprofile=coverage.out ./...
      
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  security-scan:
    name: Security Scanning
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-no-fail -fmt sarif -out results.sarif ./...'
        
    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: results.sarif

  performance-test:
    name: Performance Testing
    runs-on: ubuntu-latest
    needs: validate-generation
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
        
    - name: Install go-validate
      run: go install github.com/mateothegreat/go-config/cmd/go-validate@latest
      
    - name: Generate validation code
      run: go generate ./...
      
    - name: Run benchmarks
      run: |
        go test -bench=. -benchmem -count=3 ./... > benchmark.txt
        echo "## Benchmark Results" >> $GITHUB_STEP_SUMMARY
        echo '```' >> $GITHUB_STEP_SUMMARY
        cat benchmark.txt >> $GITHUB_STEP_SUMMARY
        echo '```' >> $GITHUB_STEP_SUMMARY
        
    - name: Performance regression check
      run: |
        # Compare with baseline performance
        go test -bench=BenchmarkValidation -count=1 ./... | \
        grep "BenchmarkValidation" | \
        awk '{if ($3 > 100) {print "❌ Performance regression detected: " $3 " ns/op"; exit 1} else {print "✅ Performance OK: " $3 " ns/op"}}'

  lint-and-format:
    name: Lint and Format
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
        
    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout 5m
        
    - name: Check formatting
      run: |
        if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
          echo "❌ Code is not formatted. Run 'gofmt -s -w .'"
          gofmt -s -l .
          exit 1
        fi
        echo "✅ Code is properly formatted"

  docker-build:
    name: Docker Build and Test
    runs-on: ubuntu-latest
    needs: [validate-generation, security-scan]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
      
    - name: Build Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./docker/Dockerfile
        push: false
        tags: myapp:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
        
    - name: Test Docker image
      run: |
        docker run --rm myapp:${{ github.sha }} --version
        docker run --rm myapp:${{ github.sha }} --validate-config
```

### 2. Deployment Workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    tags: [ 'v*' ]

jobs:
  deploy:
    name: Deploy to Production
    runs-on: ubuntu-latest
    environment: production
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        
    - name: Install go-validate
      run: go install github.com/mateothegreat/go-config/cmd/go-validate@latest
      
    - name: Generate and validate production config
      run: |
        go generate ./...
        go-validate generate --verbose
        
        # Validate production configuration
        ./scripts/validate-production-config.sh
        
    - name: Build production binary
      run: |
        CGO_ENABLED=0 GOOS=linux go build \
          -ldflags="-w -s -X main.version=${{ github.ref_name }}" \
          -o bin/app ./cmd/server
          
    - name: Login to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
        
    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./docker/Dockerfile.production
        push: true
        tags: |
          ghcr.io/${{ github.repository }}:${{ github.ref_name }}
          ghcr.io/${{ github.repository }}:latest
          
    - name: Deploy to Kubernetes
      run: |
        echo "${{ secrets.KUBECONFIG }}" | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig
        
        # Update deployment with new image
        kubectl set image deployment/myapp \
          myapp=ghcr.io/${{ github.repository }}:${{ github.ref_name }}
          
        # Wait for rollout
        kubectl rollout status deployment/myapp --timeout=600s
        
        # Verify deployment
        kubectl get pods -l app=myapp
```

## 🐳 Docker Integration

### 1. Multi-Stage Production Dockerfile

```dockerfile
# docker/Dockerfile.production
FROM golang:1.21-alpine AS builder

# Install go-validate
RUN go install github.com/mateothegreat/go-config/cmd/go-validate@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate validation code
RUN go generate ./...
RUN go-validate generate --verbose

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o bin/app ./cmd/server

# Verify the build
RUN ./bin/app --version

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Create non-root user
RUN addgroup -g 1001 appgroup && \
    adduser -D -s /bin/sh -u 1001 -G appgroup appuser

# Copy binary from builder
COPY --from=builder /app/bin/app .
COPY --from=builder /app/configs/ ./configs/

# Change ownership
RUN chown -R appuser:appgroup /root

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./app --health-check || exit 1

EXPOSE 8080

CMD ["./app"]
```

### 2. Development Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: docker/Dockerfile
      target: development
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - LOG_LEVEL=debug
    volumes:
      - .:/app
      - go-mod-cache:/go/pkg/mod
    depends_on:
      - postgres
      - redis
    command: |
      sh -c "
        go generate ./... &&
        go-validate generate --verbose &&
        go run ./cmd/server
      "

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  # Code generation service for development
  codegen:
    build:
      context: .
      dockerfile: docker/Dockerfile
      target: development
    volumes:
      - .:/app
    working_dir: /app
    command: |
      sh -c "
        while inotifywait -e modify,create,delete -r . --exclude='.*_generated\.go$$'; do
          echo 'Files changed, regenerating validation code...'
          go generate ./...
          go-validate generate --verbose
        done
      "

volumes:
  postgres-data:
  redis-data:
  go-mod-cache:
```

## 🔧 Build Scripts and Automation

### 1. Production Configuration Validation

```bash
#!/bin/bash
# scripts/validate-production-config.sh

set -e

echo "🔍 Validating production configuration..."

# Check required environment variables
required_vars=(
    "DATABASE_URL"
    "REDIS_URL"
    "JWT_SECRET"
    "API_KEY"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required environment variable $var is not set"
        exit 1
    fi
done

# Load and validate configuration
export CONFIG_FILE="configs/production.yaml"

# Generate validation code
go generate ./...
go-validate generate --verbose

# Build and run config validation
go build -o /tmp/config-validator ./cmd/config-validator
/tmp/config-validator --config "$CONFIG_FILE" --validate-only

echo "✅ Production configuration is valid"
```

### 2. Security Scanning Script

```bash
#!/bin/bash
# scripts/security-scan.sh

set -e

echo "🔒 Running security scans..."

# Install security tools
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run gosec
echo "Running gosec..."
gosec -fmt json -out gosec-report.json ./...
if [ $? -ne 0 ]; then
    echo "❌ Security issues found by gosec"
    cat gosec-report.json
    exit 1
fi

# Run vulnerability check
echo "Running govulncheck..."
govulncheck ./...
if [ $? -ne 0 ]; then
    echo "❌ Vulnerabilities found"
    exit 1
fi

# Check for secrets in code
echo "Checking for secrets..."
if grep -r -i "password\|secret\|key\|token" --include="*.go" . | grep -v "_test.go" | grep -v "// "; then
    echo "❌ Potential secrets found in code"
    exit 1
fi

echo "✅ Security scans passed"
```

### 3. Performance Baseline Script

```bash
#!/bin/bash
# scripts/performance-baseline.sh

set -e

echo "📊 Establishing performance baseline..."

# Generate validation code
go generate ./...
go-validate generate --verbose

# Run benchmarks multiple times for stability
echo "Running benchmarks..."
go test -bench=BenchmarkValidation -count=5 -benchmem ./... > benchmark-results.txt

# Extract average performance
avg_ns=$(grep "BenchmarkValidation" benchmark-results.txt | \
         awk '{sum+=$3; count++} END {print sum/count}')

echo "Average validation time: ${avg_ns} ns/op"

# Set baseline (fail if significantly slower than expected)
baseline_ns=50  # 50 ns/op baseline
if (( $(echo "$avg_ns > $baseline_ns * 2" | bc -l) )); then
    echo "❌ Performance regression: ${avg_ns} ns/op exceeds baseline of ${baseline_ns} ns/op"
    exit 1
fi

echo "✅ Performance within acceptable range"

# Store results for trending
echo "${avg_ns}" > performance-baseline.txt
git add performance-baseline.txt
```

## 🔒 Security and Compliance

### 1. Secrets Management

```yaml
# .github/workflows/secrets.yml
name: Secrets Management

on:
  workflow_call:
    secrets:
      DATABASE_URL:
        required: true
      JWT_SECRET:
        required: true
      API_KEYS:
        required: true

jobs:
  validate-secrets:
    runs-on: ubuntu-latest
    steps:
    - name: Validate secret formats
      run: |
        # Validate JWT secret length
        if [ ${#JWT_SECRET} -lt 32 ]; then
          echo "❌ JWT_SECRET must be at least 32 characters"
          exit 1
        fi
        
        # Validate database URL format
        if [[ ! "$DATABASE_URL" =~ ^postgres:// ]]; then
          echo "❌ DATABASE_URL must be a valid PostgreSQL URL"
          exit 1
        fi
        
        echo "✅ All secrets are valid"
      env:
        JWT_SECRET: ${{ secrets.JWT_SECRET }}
        DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

### 2. Compliance Checks

```bash
#!/bin/bash
# scripts/compliance-check.sh

set -e

echo "🛡️  Running compliance checks..."

# Check for required license headers
find . -name "*.go" -not -path "./vendor/*" | while read file; do
    if ! head -5 "$file" | grep -q "Copyright\|License"; then
        echo "❌ Missing license header in $file"
        exit 1
    fi
done

# Check for proper error handling
if grep -r "panic\|os.Exit" --include="*.go" . | grep -v "_test.go" | grep -v "main.go"; then
    echo "❌ Found panic or os.Exit in non-main code"
    exit 1
fi

# Validate configuration schema
go run ./cmd/schema-validator --config-dir ./configs

echo "✅ Compliance checks passed"
```

## 📊 Monitoring and Observability

### 1. Application Metrics

```go
// cmd/server/metrics.go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    configValidationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "config_validation_duration_seconds",
            Help: "Time spent validating configuration",
        },
        []string{"config_type", "status"},
    )
    
    configValidationErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "config_validation_errors_total",
            Help: "Total number of configuration validation errors",
        },
        []string{"field", "error_code"},
    )
)

func validateWithMetrics(config interface{}) error {
    start := time.Now()
    
    err := validateConfig(config)
    
    status := "success"
    if err != nil {
        status = "error"
        
        // Record specific validation errors
        if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
            for _, verr := range validationErrors.Errors() {
                configValidationErrors.WithLabelValues(verr.Field, verr.Code).Inc()
            }
        }
    }
    
    configValidationDuration.WithLabelValues(
        reflect.TypeOf(config).Name(),
        status,
    ).Observe(time.Since(start).Seconds())
    
    return err
}
```

### 2. Health Check Endpoint

```go
// cmd/server/health.go
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().Unix(),
        "version":   version,
    }
    
    // Validate current configuration
    if err := validateCurrentConfig(); err != nil {
        health["status"] = "unhealthy"
        health["config_errors"] = formatValidationErrors(err)
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    json.NewEncoder(w).Encode(health)
}
```

## 🎓 Best Practices for Production

### 1. Configuration Management

- **Environment-specific configs** - Separate configs for dev/staging/prod
- **Secret management** - Use external secret stores (Vault, K8s secrets)
- **Configuration validation** - Validate configs before deployment
- **Zero-downtime updates** - Rolling updates with health checks

### 2. CI/CD Pipeline Design

- **Fast feedback** - Parallel jobs for quick validation
- **Security first** - Scan before deployment
- **Comprehensive testing** - Unit, integration, and performance tests
- **Automated rollback** - Automatic rollback on failure

### 3. Monitoring Strategy

- **Validation metrics** - Track validation performance and errors
- **Health checks** - Continuous configuration validation
- **Alerting** - Alert on configuration validation failures
- **Observability** - Structured logging and tracing

## 🎓 Next Steps

- Explore performance testing in [`../13-performance-testing/`](../examples/13-performance-testing)
- Learn build automation in [`../12-build-automation/`](../examples/12-build-automation)
- See error handling patterns in [`../10-error-handling/`](../examples/10-error-handling)
