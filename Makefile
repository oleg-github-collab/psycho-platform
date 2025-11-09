.PHONY: help build run test clean docker-up docker-down migrate

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/api cmd/api/main.go
	@echo "✓ Build complete"

run: ## Run the application
	@echo "Starting application..."
	@go run cmd/api/main.go

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...
	@echo "✓ Tests complete"

test-api: ## Test API endpoints
	@echo "Testing API..."
	@bash test/api_test.sh

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf uploads/
	@echo "✓ Clean complete"

docker-up: ## Start Docker services (PostgreSQL + Redis)
	@echo "Starting Docker services..."
	@docker run -d --name psycho-postgres \
		-e POSTGRES_DB=psycho_platform \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-p 5432:5432 \
		postgres:15 || echo "PostgreSQL already running"
	@docker run -d --name psycho-redis \
		-p 6379:6379 \
		redis:7 || echo "Redis already running"
	@sleep 2
	@echo "✓ Docker services started"

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	@docker stop psycho-postgres psycho-redis || true
	@docker rm psycho-postgres psycho-redis || true
	@echo "✓ Docker services stopped"

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run cmd/api/main.go migrate
	@echo "✓ Migrations complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

dev: docker-up ## Start development environment
	@echo "Starting development environment..."
	@sleep 3
	@make run

check: ## Check code quality
	@echo "Running checks..."
	@go fmt ./...
	@go vet ./...
	@echo "✓ Checks complete"

railway-deploy: ## Deploy to Railway
	@echo "Deploying to Railway..."
	@railway up
	@echo "✓ Deployed"

all: clean deps build ## Build everything
	@echo "✓ Build complete"
