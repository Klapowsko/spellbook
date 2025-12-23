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
)

// MockGeminiServiceTopics é um mock do GeminiService para testes de topics
type MockGeminiServiceTopics struct {
	mock.Mock
}

func (m *MockGeminiServiceTopics) GenerateRoadmap(topic string) (*models.Roadmap, error) {
	args := m.Called(topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Roadmap), args.Error(1)
}

func (m *MockGeminiServiceTopics) GenerateTopics(subject string, count int) (*models.TopicsResponse, error) {
	args := m.Called(subject, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TopicsResponse), args.Error(1)
}

func TestTopicsHandler_GenerateTopics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockGeminiServiceTopics)
	handler := &TopicsHandler{GeminiService: mockService}

	expectedTopics := &models.TopicsResponse{
		Subject: "Python",
		Topics:  []string{"OOP", "Decorators", "Context Managers"},
	}

	mockService.On("GenerateTopics", "Python", 10).Return(expectedTopics, nil)

	router := gin.New()
	router.POST("/topics", handler.GenerateTopics)

	reqBody := models.TopicsRequest{Subject: "Python", Count: 10}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/topics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TopicsResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Python", response.Subject)
	assert.Len(t, response.Topics, 3)

	mockService.AssertExpectations(t)
}

func TestTopicsHandler_GenerateTopics_EmptySubject(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockGeminiServiceTopics)
	handler := &TopicsHandler{GeminiService: mockService}

	router := gin.New()
	router.POST("/topics", handler.GenerateTopics)

	reqBody := models.TopicsRequest{Subject: ""}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/topics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTopicsHandler_GenerateTopics_DefaultCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockGeminiServiceTopics)
	handler := &TopicsHandler{GeminiService: mockService}

	expectedTopics := &models.TopicsResponse{
		Subject: "JavaScript",
		Topics:  []string{"Closures", "Promises", "Async/Await"},
	}

	// Quando count não é especificado, deve usar 10 como default
	mockService.On("GenerateTopics", "JavaScript", 10).Return(expectedTopics, nil)

	router := gin.New()
	router.POST("/topics", handler.GenerateTopics)

	reqBody := models.TopicsRequest{Subject: "JavaScript", Count: 0}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/topics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

