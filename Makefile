.PHONY: help build run test test-bdd test-unit clean docker-build docker-run docker-compose-up docker-compose-down

# Variáveis
APP_NAME=spellbook
BINARY_NAME=spellbook
DOCKER_IMAGE=spellbook:latest
DOCKER_CONTAINER=spellbook-container
GO_VERSION=1.21

help: ## Mostra esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compila a aplicação
	@echo "Compilando $(APP_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/server/main.go
	@echo "Build concluído: bin/$(BINARY_NAME)"

run: ## Executa a aplicação localmente
	@echo "Executando $(APP_NAME)..."
	@go run ./cmd/server/main.go

test: test-unit test-bdd ## Executa todos os testes

test-unit: ## Executa testes unitários
	@echo "Executando testes unitários..."
	@go test -v ./internal/...

test-bdd: ## Executa testes BDD (Godog)
	@echo "Executando testes BDD..."
	@go test -v ./features/...

test-coverage: ## Executa testes com cobertura
	@echo "Executando testes com cobertura..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório de cobertura gerado: coverage.html"

clean: ## Limpa arquivos gerados
	@echo "Limpando arquivos..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

deps: ## Instala dependências
	@echo "Instalando dependências..."
	@go mod download
	@go mod tidy

docker-build: ## Constrói a imagem Docker
	@echo "Construindo imagem Docker..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Imagem $(DOCKER_IMAGE) construída com sucesso"

docker-run: ## Executa container Docker
	@echo "Executando container Docker..."
	@docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-compose-up: ## Inicia serviços com docker-compose
	@echo "Iniciando serviços com docker-compose..."
	@docker-compose up -d
	@echo "Serviços iniciados. Use 'make docker-compose-logs' para ver os logs"

docker-compose-down: ## Para serviços do docker-compose
	@echo "Parando serviços do docker-compose..."
	@docker-compose down

docker-compose-logs: ## Mostra logs do docker-compose
	@docker-compose logs -f

docker-compose-restart: ## Reinicia serviços do docker-compose
	@docker-compose restart

lint: ## Executa linter
	@echo "Executando linter..."
	@golangci-lint run ./... || echo "golangci-lint não instalado. Instale com: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

fmt: ## Formata o código
	@echo "Formatando código..."
	@go fmt ./...

vet: ## Executa go vet
	@echo "Executando go vet..."
	@go vet ./...

