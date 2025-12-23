# Build stage
FROM golang:1.25-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/spellbook ./cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Instalar ca-certificates e curl para HTTPS e healthcheck
RUN apk --no-cache add ca-certificates tzdata curl

# Criar usuário não-root
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copiar binário do stage de build
COPY --from=builder /app/bin/spellbook .

# Mudar propriedade para usuário não-root
RUN chown -R appuser:appuser /app

# Mudar para usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Comando para executar a aplicação
CMD ["./spellbook"]

