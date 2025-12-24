# Makefile for Task Management API Project

.PHONY: help up down build logs test seed dev clean restart frontend-logs api-logs db-logs code-cov

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

up: ## Start all services in detached mode
	docker compose up -d

down: ## Stop all services
	docker compose down

build: ## Build all services
	docker compose build

logs: ## Show logs for all services
	docker compose logs -f

test: ## Run tests in the API container
	docker compose exec api go test ./...

seed: ## Run the database seeder
	docker compose --profile seeder up seeder

dev: ## Start services in development mode (with live reload)
	docker compose up

clean: ## Stop services and remove volumes
	docker compose down -v

restart: ## Restart all services
	docker compose restart

frontend-logs: ## Show logs for frontend service
	docker compose logs -f frontend

api-logs: ## Show logs for API service
	docker compose logs -f api

db-logs: ## Show logs for database service
	docker compose logs -f postgres

code-cov: ## Generate code coverage report
	go test ./... -coverprofile=coverage.out
	@echo "Coverage Summary:"
	go tool cover -func=coverage.out
	@coverage_line=$$(go tool cover -func=coverage.out | grep total); \
	coverage_percent=$$(echo "$$coverage_line" | grep -o '[0-9]\+\.[0-9]\+' | head -1); \
	if [ -n "$$coverage_percent" ]; then \
		coverage_num=$$(echo "$$coverage_percent" | cut -d. -f1); \
		if [ "$$coverage_num" -ge 70 ]; then color='\033[0;32m'; \
		elif [ "$$coverage_num" -ge 50 ]; then color='\033[1;33m'; \
		else color='\033[0;31m'; fi; \
		echo "Total Coverage: $${color}$${coverage_percent}%\033[0m"; \
	fi
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ HTML report generated: coverage.html"

prometheus-logs: ## Show logs for Prometheus service
	docker compose logs -f prometheus