.PHONY: help build up down logs restart clean test gpu-check

# Variables
COMPOSE_FILE := docker-compose.yml
COMPOSE_INTEGRATED := docker-compose.integrated.yml
SERVICE_NAME := transcription-service3
IMAGE_NAME := recontext-transcription-service3

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the Docker image
	docker-compose -f $(COMPOSE_FILE) build

build-integrated: ## Build the Docker image for integrated mode
	docker-compose -f $(COMPOSE_INTEGRATED) build

up: ## Start all services (standalone mode)
	docker-compose -f $(COMPOSE_FILE) up -d

up-integrated: ## Start transcription service only (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) up -d

down: ## Stop and remove containers
	docker-compose -f $(COMPOSE_FILE) down

down-integrated: ## Stop transcription service (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) down

logs: ## View logs
	docker-compose -f $(COMPOSE_FILE) logs -f $(SERVICE_NAME)

logs-integrated: ## View logs (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) logs -f $(SERVICE_NAME)

restart: ## Restart the transcription service
	docker-compose -f $(COMPOSE_FILE) restart $(SERVICE_NAME)

restart-integrated: ## Restart transcription service (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) restart $(SERVICE_NAME)

clean: ## Remove containers, volumes, and images
	docker-compose -f $(COMPOSE_FILE) down -v
	docker rmi $(IMAGE_NAME):latest || true

clean-integrated: ## Remove containers and volumes (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) down -v
	docker rmi sivanov2018/$(IMAGE_NAME):latest || true

ps: ## Show running containers
	docker-compose -f $(COMPOSE_FILE) ps

ps-integrated: ## Show running containers (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) ps

shell: ## Open a shell in the transcription service container
	docker exec -it recontext-$(SERVICE_NAME) /bin/bash

gpu-check: ## Check if GPU is available in container
	docker run --rm --gpus all nvcr.io/nvidia/pytorch:25.04-py3 nvidia-smi

test-env: ## Test environment configuration
	docker-compose -f $(COMPOSE_FILE) config

test-env-integrated: ## Test environment configuration (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) config

stats: ## Show resource usage statistics
	docker stats recontext-$(SERVICE_NAME) --no-stream

health: ## Check health status
	docker inspect --format='{{json .State.Health}}' recontext-$(SERVICE_NAME) | jq

pull: ## Pull the base image
	docker pull nvcr.io/nvidia/pytorch:25.04-py3

push: ## Push image to registry (integrated mode)
	docker-compose -f $(COMPOSE_INTEGRATED) build
	docker tag recontext-$(SERVICE_NAME):latest sivanov2018/recontext-$(SERVICE_NAME):latest
	docker push sivanov2018/recontext-$(SERVICE_NAME):latest

dev: ## Run service locally (outside Docker)
	python main.py

install: ## Install Python dependencies locally
	pip install -r requirements.txt

# Quick commands
start: up ## Alias for 'up'
stop: down ## Alias for 'down'
