package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/spellbook/spellbook/internal/models"
	"github.com/spellbook/spellbook/internal/services"
)

// MockGeminiService é um mock do GeminiService para testes
// Garantir que implementa a interface GeminiServiceInterface
var _ services.GeminiServiceInterface = (*MockGeminiService)(nil)

type MockGeminiService struct {
	mock.Mock
}

func (m *MockGeminiService) GenerateRoadmap(topic string) (*models.Roadmap, error) {
	args := m.Called(topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Roadmap), args.Error(1)
}

func (m *MockGeminiService) GenerateTopics(subject string, count int) (*models.TopicsResponse, error) {
	args := m.Called(subject, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TopicsResponse), args.Error(1)
}

func TestRoadmapHandler_GenerateRoadmap_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockGeminiService)
	handler := &RoadmapHandler{GeminiService: mockService}

	expectedRoadmap := &models.Roadmap{
		Topic: "Machine Learning",
		Roadmap: []models.RoadmapCategory{
			{
				Category: "Fundamentos",
				Items: []models.RoadmapItem{
					{ID: "1", Title: "Introdução", Completed: false},
				},
			},
		},
	}

	mockService.On("GenerateRoadmap", "Machine Learning").Return(expectedRoadmap, nil)

	router := gin.New()
	router.POST("/roadmap", handler.GenerateRoadmap)

	reqBody := models.RoadmapRequest{Topic: "Machine Learning"}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/roadmap", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Roadmap
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Machine Learning", response.Topic)
	assert.Len(t, response.Roadmap, 1)

	mockService.AssertExpectations(t)
}

func TestRoadmapHandler_GenerateRoadmap_EmptyTopic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockGeminiService)
	handler := &RoadmapHandler{GeminiService: mockService}

	router := gin.New()
	router.POST("/roadmap", handler.GenerateRoadmap)

	reqBody := models.RoadmapRequest{Topic: ""}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/roadmap", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoadmapHandler_GenerateRoadmap_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockGeminiService)
	handler := &RoadmapHandler{GeminiService: mockService}

	mockService.On("GenerateRoadmap", "Test").Return(nil, assert.AnError)

	router := gin.New()
	router.POST("/roadmap", handler.GenerateRoadmap)

	reqBody := models.RoadmapRequest{Topic: "Test"}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/roadmap", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

