package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeminiService_NewGeminiService(t *testing.T) {
	apiKey := "test-api-key"
	service := NewGeminiService(apiKey)

	assert.NotNil(t, service)
	assert.Equal(t, apiKey, service.APIKey)
	assert.NotNil(t, service.HTTPClient)
	assert.Equal(t, "https://generativelanguage.googleapis.com/v1beta", service.BaseURL)
}

func TestGeminiService_GenerateRoadmap_EmptyTopic(t *testing.T) {
	service := NewGeminiService("test-key")
	
	roadmap, err := service.GenerateRoadmap("")
	
	assert.Nil(t, roadmap)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tópico não pode ser vazio")
}

func TestGeminiService_GenerateTopics_EmptySubject(t *testing.T) {
	service := NewGeminiService("test-key")
	
	topics, err := service.GenerateTopics("", 10)
	
	assert.Nil(t, topics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "assunto não pode ser vazio")
}

func TestGeminiService_GenerateTopics_DefaultCount(t *testing.T) {
	service := NewGeminiService("test-key")
	
	// Testa que count 0 ou negativo usa default
	// Como não temos API key real, vamos apenas testar a validação
	topics, err := service.GenerateTopics("Python", 0)
	
	// Deve falhar por falta de API key, mas não por count inválido
	assert.Nil(t, topics)
	assert.Error(t, err)
	// O erro não deve ser sobre count
	assert.NotContains(t, err.Error(), "count")
}

func TestCleanJSONText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON com markdown code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "JSON sem markdown",
			input:    "{\"key\": \"value\"}",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "JSON com texto antes",
			input:    "Aqui está o JSON: {\"key\": \"value\"}",
			expected: "{\"key\": \"value\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSONText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

