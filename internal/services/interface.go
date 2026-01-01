package services

import "github.com/spellbook/spellbook/internal/models"

// GeminiServiceInterface define a interface para o serviço Gemini
// Isso permite criar mocks para testes
type GeminiServiceInterface interface {
	GenerateRoadmap(topic string, availableDays *int, exactItemCount *int) (*models.Roadmap, error)
	GenerateTopics(subject string, count int) (*models.TopicsResponse, error)
	GenerateKeyResults(objective string, count int, completionDate *string) (*models.KeyResultsResponse, error)
	GenerateEducationalRoadmap(topic string) (*models.EducationalRoadmap, error)
	GenerateEducationalTrail(topic string, availableDays *int) (*models.EducationalTrail, error)
}

