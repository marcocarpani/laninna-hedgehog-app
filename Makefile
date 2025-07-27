# 🦔 La Ninna - Hedgehog Management System Makefile

.PHONY: help build run test clean dev docker install deps fmt vet lint

# Default target
help: ## Show this help message
	@echo "🦔 La Ninna - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development
dev: ## Run with hot reload (requires air)
	@echo "🔥 Starting development server with hot reload..."
	air

run: ## Run the application
	@echo "🚀 Starting La Ninna server..."
	go run .

install: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

deps: install ## Alias for install

# Building
build: ## Build the application
	@echo "🔨 Building La Ninna..."
	go build -o bin/laninna-app .

build-linux: ## Build for Linux
	@echo "🐧 Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/laninna-linux .

build-windows: ## Build for Windows
	@echo "🪟 Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o bin/laninna.exe .

build-mac: ## Build for macOS
	@echo "🍎 Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o bin/laninna-mac .

build-all: build-linux build-windows build-mac ## Build for all platforms

# Testing
test: ## Run tests
	@echo "🧪 Running tests..."
	go test ./...

test-verbose: ## Run tests with verbose output
	@echo "🧪 Running tests (verbose)..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	go test -cover ./...

test-coverage-html: ## Generate HTML coverage report
	@echo "📊 Generating HTML coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Code Quality
fmt: ## Format code
	@echo "✨ Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	go vet ./...

lint: ## Run golangci-lint (requires golangci-lint)
	@echo "🔍 Running linter..."
	golangci-lint run

check: fmt vet ## Run formatting and vetting

# Database
db-reset: ## Reset database (removes laninna.db)
	@echo "🗑️ Resetting database..."
	rm -f laninna.db laninna.db-wal laninna.db-shm

db-backup: ## Backup database
	@echo "💾 Backing up database..."
	cp laninna.db laninna-backup-$(shell date +%Y%m%d-%H%M%S).db

db-inspect: ## Open database in sqlite3
	@echo "🔍 Opening database..."
	sqlite3 laninna.db

# Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t laninna-app .

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/data:/root laninna-app

docker-dev: ## Run Docker container with volume mount for development
	@echo "🐳 Running Docker container (development)..."
	docker run -p 8080:8080 -v $(PWD):/app -w /app golang:1.19 go run .

# Deployment
deploy-build: ## Build optimized binary for deployment
	@echo "🚀 Building optimized binary..."
	CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/laninna-app .

deploy-package: deploy-build ## Package for deployment
	@echo "📦 Packaging for deployment..."
	mkdir -p deploy
	cp bin/laninna-app deploy/
	cp -r templates deploy/
	cp -r static deploy/
	tar -czf laninna-deploy.tar.gz -C deploy .
	@echo "Deployment package: laninna-deploy.tar.gz"

# Maintenance
clean: ## Clean build artifacts
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -rf deploy/
	rm -f laninna-deploy.tar.gz
	rm -f coverage.out coverage.html
	go clean

clean-all: clean db-reset ## Clean everything including database

logs: ## Show application logs (if running with systemd)
	journalctl -u laninna-app -f

# Development Tools
install-tools: ## Install development tools
	@echo "🛠️ Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

setup: install install-tools ## Complete development setup

# API Testing
api-test: ## Test API endpoints
	@echo "🧪 Testing API endpoints..."
	@echo "Login test:"
	curl -s -X POST http://localhost:8080/api/login \
		-H "Content-Type: application/json" \
		-d '{"username":"admin","password":"admin123"}' | jq .
	@echo "\nHealth check:"
	curl -s http://localhost:8080/health | jq .

# Performance
benchmark: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem ./...

profile: ## Run with profiling
	@echo "📊 Starting with profiling..."
	go run . -cpuprofile=cpu.prof -memprofile=mem.prof

# Security
security-scan: ## Run security scan (requires gosec)
	@echo "🔒 Running security scan..."
	gosec ./...

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	godoc -http=:6060
	@echo "Documentation available at: http://localhost:6060"

swagger: ## Generate Swagger documentation
	@echo "📚 Generating Swagger docs..."
	swag init
	@echo "Swagger UI available at: http://localhost:8080/swagger/index.html"

# Git hooks
git-hooks: ## Install git hooks
	@echo "🪝 Installing git hooks..."
	cp scripts/pre-commit .git/hooks/
	chmod +x .git/hooks/pre-commit

# Quick commands
start: run ## Alias for run
stop: ## Stop running processes
	@echo "🛑 Stopping processes..."
	pkill -f "laninna-app" || true

restart: stop start ## Restart application

status: ## Check if application is running
	@echo "📊 Application status:"
	@pgrep -f "laninna-app" > /dev/null && echo "✅ Running" || echo "❌ Not running"
	@curl -s http://localhost:8080/health > /dev/null && echo "✅ Health check OK" || echo "❌ Health check failed"

# Release
version: ## Show version info
	@echo "🦔 La Ninna Version Info:"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $(shell date)"

release: clean test build-all deploy-package ## Create release package
	@echo "🎉 Release package created!"

# Default development workflow
dev-setup: setup db-reset run ## Complete development setup and run

# Production workflow  
prod-deploy: test deploy-build ## Production deployment build

# CI/CD
ci: deps fmt vet test swagger ## CI pipeline
	@echo "✅ CI pipeline completed successfully"

cd: ci build ## CD pipeline
	@echo "✅ CD pipeline completed successfully"