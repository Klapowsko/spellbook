.PHONY: help build up down logs restart test

# Variáveis
DOCKER_IMAGE=spellbook:latest
CONTAINER_NAME=spellbook-api

help: ## Mostra esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Constrói a imagem Docker
	@docker build -t $(DOCKER_IMAGE) .

rebuild: ## Reconstrói a imagem Docker sem cache
	@docker build --no-cache -t $(DOCKER_IMAGE) .

up: ## Inicia serviços com docker-compose
	@docker compose up -d

up-build: ## Reconstrói e inicia serviços (força rebuild sem cache)
	@docker compose build --no-cache
	@docker compose up -d

down: ## Para serviços do docker-compose
	@docker compose down

logs: ## Mostra logs do docker-compose
	@docker compose logs -f

restart: ## Reinicia serviços do docker-compose
	@docker compose restart

test: ## Executa testes dentro do container
	@docker compose exec spellbook go test -v ./...
