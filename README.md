# Spellbook - Serviço Centralizado de IA

Serviço de back-end em Go que centraliza todas as integrações com Inteligência Artificial, começando com a integração do **Gemini Flash (versão free)**.

## 🎯 Funcionalidades

- **POST /roadmap** - Gera roadmaps de estudo estruturados usando IA
- **POST /topics** - Gera lista de tópicos sobre um assunto

## 🚀 Instalação

### Pré-requisitos

- Go 1.21 ou superior
- API Key do Google Gemini ([obter aqui](https://makersuite.google.com/app/apikey))

### Configuração

1. Clone o repositório:
```bash
git clone <repo-url>
cd spellbook
```

2. Instale as dependências:
```bash
go mod download
```

3. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env e adicione sua GEMINI_API_KEY
```

## 🏃 Executando

### Usando Makefile (Recomendado)

```bash
# Ver todos os comandos disponíveis
make help

# Executar localmente
make run

# Compilar
make build

# Executar testes
make test

# Executar testes unitários
make test-unit

# Executar testes BDD
make test-bdd
```

### Executando Manualmente

#### Servidor de Desenvolvimento
```bash
go run cmd/server/main.go
```

O servidor estará disponível em `http://localhost:8080`

#### Executar Testes BDD (Godog)
```bash
godog features/
```

#### Executar Testes Unitários
```bash
go test ./...
```

## 🐳 Docker

### Usando Docker Compose (Recomendado)

```bash
# Iniciar serviços
make docker-compose-up

# Ver logs
make docker-compose-logs

# Parar serviços
make docker-compose-down

# Reiniciar serviços
make docker-compose-restart
```

### Usando Docker diretamente

```bash
# Construir imagem
make docker-build
# ou
docker build -t spellbook:latest .

# Executar container
make docker-run
# ou
docker run --rm -p 8080:8080 --env-file .env spellbook:latest
```

### Variáveis de Ambiente no Docker

Certifique-se de ter um arquivo `.env` com:
```
GEMINI_API_KEY=sua_chave_aqui
PORT=8080
```

## 📚 API

### POST /roadmap

Gera um roadmap de estudo estruturado sobre um tópico.

**Request:**
```json
{
  "topic": "Machine Learning"
}
```

**Response:**
```json
{
  "topic": "Machine Learning",
  "roadmap": [
    {
      "category": "Fundamentos",
      "items": [
        {"id": "1", "title": "Introdução à ML", "completed": false},
        {"id": "2", "title": "Estatística básica", "completed": false}
      ]
    }
  ]
}
```

### POST /topics

Gera uma lista de tópicos relacionados a um assunto.

**Request:**
```json
{
  "subject": "Python",
  "count": 10
}
```

**Response:**
```json
{
  "subject": "Python",
  "topics": [
    "Programação Orientada a Objetos",
    "Decorators",
    "Context Managers"
  ]
}
```

## 🧪 Metodologia de Desenvolvimento

Este projeto segue uma abordagem **BDD primeiro, depois TDD**:

1. **BDD**: Features do Godog descrevendo comportamentos esperados
2. **TDD**: Testes unitários baseados nas features
3. **Implementação**: Código para fazer os testes passarem
4. **Refatoração**: Melhorias mantendo testes passando

## 📁 Estrutura do Projeto

```
spellbook/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── app/                     # Inicialização da aplicação
│   ├── handlers/                # Handlers HTTP
│   ├── services/                # Lógica de negócio
│   ├── models/                  # Estruturas de dados
│   ├── config/                  # Configuração
│   ├── middleware/              # Middlewares (CORS, etc)
│   └── routes/                  # Configuração de rotas
├── features/                     # Testes BDD (Godog)
│   └── step_definitions/        # Step definitions
├── bin/                          # Binários compilados
├── Dockerfile                    # Configuração Docker
├── docker-compose.yml            # Orquestração Docker
├── Makefile                      # Comandos automatizados
└── go.mod
```

## 🔧 Tecnologias

- **Go (Golang)** - Linguagem principal
- **Gin** - Framework web
- **Godog** - Framework BDD (Cucumber para Go)
- **Testify** - Biblioteca de assertions
- **Google Gemini API** - Integração com IA

## 📝 Licença

MIT

