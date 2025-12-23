#!/bin/bash

echo "=== Verificando Deploy do Spellbook ==="
echo ""

echo "1. Verificando se o arquivo educational_roadmap.go existe:"
if [ -f "internal/models/educational_roadmap.go" ]; then
    echo "   ✓ Arquivo existe"
else
    echo "   ✗ Arquivo NÃO existe - precisa fazer pull/copiar"
    exit 1
fi

echo ""
echo "2. Verificando se o handler tem o método GenerateEducationalRoadmap:"
if grep -q "GenerateEducationalRoadmap" internal/handlers/roadmap.go; then
    echo "   ✓ Método encontrado no handler"
else
    echo "   ✗ Método NÃO encontrado - arquivo não atualizado"
    exit 1
fi

echo ""
echo "3. Verificando se a rota está registrada:"
if grep -q "educational-roadmap" internal/routes/routes.go; then
    echo "   ✓ Rota encontrada"
else
    echo "   ✗ Rota NÃO encontrada - arquivo não atualizado"
    exit 1
fi

echo ""
echo "4. Verificando se o serviço Gemini tem o método:"
if grep -q "GenerateEducationalRoadmap" internal/services/gemini.go; then
    echo "   ✓ Método encontrado no serviço"
else
    echo "   ✗ Método NÃO encontrado - arquivo não atualizado"
    exit 1
fi

echo ""
echo "5. Testando compilação:"
if go build -o /tmp/spellbook-test ./cmd/server/main.go 2>&1; then
    echo "   ✓ Compilação bem-sucedida"
    rm -f /tmp/spellbook-test
else
    echo "   ✗ Erro na compilação"
    exit 1
fi

echo ""
echo "6. Verificando se o container está rodando:"
if docker ps | grep -q spellbook-api; then
    echo "   ✓ Container está rodando"
    CONTAINER_ID=$(docker ps | grep spellbook-api | awk '{print $1}')
    echo "   Container ID: $CONTAINER_ID"
else
    echo "   ⚠ Container não está rodando"
fi

echo ""
echo "7. Verificando imagem Docker:"
if docker images | grep -q spellbook; then
    echo "   ✓ Imagem existe"
    IMAGE_DATE=$(docker images spellbook:latest --format "{{.CreatedAt}}" | head -1)
    echo "   Data da imagem: $IMAGE_DATE"
else
    echo "   ⚠ Imagem não encontrada"
fi

echo ""
echo "=== Verificação concluída ==="
echo ""
echo "Se tudo estiver OK, execute:"
echo "  make down"
echo "  make build"
echo "  make up"
echo ""
echo "Depois teste:"
echo "  curl -X POST http://localhost:8082/api/v1/educational-roadmap -H 'Content-Type: application/json' -d '{\"topic\": \"SOLID\"}'"

