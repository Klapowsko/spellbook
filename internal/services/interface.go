package services

import "github.com/spellbook/spellbook/internal/models"

// GeminiServiceInterface define a interface para o serviço Gemini
// Isso permite criar mocks para testes
type GeminiServiceInterface interface {
	GenerateRoadmap(topic string) (*models.Roadmap, error)
	GenerateTopics(subject string, count int) (*models.TopicsResponse, error)
	GenerateEducationalRoadmap(topic string) (*models.EducationalRoadmap, error)
	GenerateEducationalTrail(topic string) (*models.EducationalTrail, error)
}

