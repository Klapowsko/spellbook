package step_definitions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/spellbook/spellbook/internal/handlers"
	"github.com/spellbook/spellbook/internal/services"
)

type topicsFeature struct {
	router   *gin.Engine
	response *httptest.ResponseRecorder
}

func (t *topicsFeature) resetResponse(*godog.Scenario) {
	gin.SetMode(gin.TestMode)
	t.router = gin.New()
	t.response = httptest.NewRecorder()
}

func (t *topicsFeature) iSendAPOSTRequestToTopicsWithSubjectAndCount(subject string, count int) error {
	geminiService := services.NewGeminiService(os.Getenv("GEMINI_API_KEY"))
	topicsHandler := handlers.NewTopicsHandler(geminiService)
	
	t.router.POST("/topics", topicsHandler.GenerateTopics)

	reqBody := map[string]interface{}{
		"subject": subject,
		"count":   count,
	}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/topics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	t.router.ServeHTTP(t.response, req)
	return nil
}

func (t *topicsFeature) iSendAPOSTRequestToTopicsWithSubject(subject string) error {
	geminiService := services.NewGeminiService(os.Getenv("GEMINI_API_KEY"))
	topicsHandler := handlers.NewTopicsHandler(geminiService)
	
	t.router.POST("/topics", topicsHandler.GenerateTopics)

	reqBody := map[string]string{"subject": subject}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/topics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	t.router.ServeHTTP(t.response, req)
	return nil
}

func (t *topicsFeature) theResponseShouldContainAListOfTopics() error {
	var topicsResp map[string]interface{}
	if err := json.Unmarshal(t.response.Body.Bytes(), &topicsResp); err != nil {
		return fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	topics, ok := topicsResp["topics"].([]interface{})
	if !ok {
		return fmt.Errorf("topics não é um array")
	}

	if len(topics) == 0 {
		return fmt.Errorf("lista de tópicos está vazia")
	}

	return nil
}

func (t *topicsFeature) theListShouldHaveAtLeastTopics(minTopics int) error {
	var topicsResp map[string]interface{}
	if err := json.Unmarshal(t.response.Body.Bytes(), &topicsResp); err != nil {
		return fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	topics, ok := topicsResp["topics"].([]interface{})
	if !ok {
		return fmt.Errorf("topics não é um array")
	}

	if len(topics) < minTopics {
		return fmt.Errorf("esperado pelo menos %d tópicos, mas recebeu %d", minTopics, len(topics))
	}

	return nil
}

func (t *topicsFeature) theSubjectShouldBe(subject string) error {
	var topicsResp map[string]interface{}
	if err := json.Unmarshal(t.response.Body.Bytes(), &topicsResp); err != nil {
		return fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	respSubject, ok := topicsResp["subject"].(string)
	if !ok || respSubject != subject {
		return fmt.Errorf("esperado subject '%s', mas recebeu '%v'", subject, topicsResp["subject"])
	}

	return nil
}

func InitializeTopicsScenario(ctx *godog.ScenarioContext) {
	topics := &topicsFeature{}

	ctx.BeforeScenario(topics.resetResponse)

	ctx.Step(`^eu envio uma requisição POST para /topics com subject "([^"]*)" e count (\d+)$`, topics.iSendAPOSTRequestToTopicsWithSubjectAndCount)
	ctx.Step(`^eu envio uma requisição POST para /topics com subject "([^"]*)"$`, topics.iSendAPOSTRequestToTopicsWithSubject)
	ctx.Step(`^a resposta deve conter uma lista de tópicos$`, topics.theResponseShouldContainAListOfTopics)
	ctx.Step(`^a lista deve ter pelo menos (\d+) tópicos$`, topics.theListShouldHaveAtLeastTopics)
	ctx.Step(`^o subject deve ser "([^"]*)"$`, topics.theSubjectShouldBe)
}

