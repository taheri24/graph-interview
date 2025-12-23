# Makefile for Task Management API Project

.PHONY: help up down build logs test seed dev clean restart frontend-logs api-logs db-logs

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

up: ## Start all services in detached mode
	docker-compose up -d

down: ## Stop all services
	docker-compose down

build: ## Build all services
	docker-compose build

logs: ## Show logs for all services
	docker-compose logs -f

test: ## Run tests in the API container
	docker-compose exec api go test ./...

seed: ## Run the database seeder
	docker-compose --profile seeder up seeder

dev: ## Start services in development mode (with live reload)
	docker-compose up

clean: ## Stop services and remove volumes
	docker-compose down -v

restart: ## Restart all services
	docker-compose restart

frontend-logs: ## Show logs for frontend service
	docker-compose logs -f frontend

api-logs: ## Show logs for API service
	docker-compose logs -f api

db-logs: ## Show logs for database service
	docker-compose logs -f postgres

prometheus-logs: ## Show logs for Prometheus service
	docker-compose logs -f prometheus