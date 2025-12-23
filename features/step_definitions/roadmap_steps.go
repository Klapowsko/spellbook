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
	"github.com/spellbook/spellbook/internal/config"
	"github.com/spellbook/spellbook/internal/handlers"
	"github.com/spellbook/spellbook/internal/services"
)

type apiFeature struct {
	router     *gin.Engine
	response   *httptest.ResponseRecorder
	apiKey     string
	originalKey string
}

func (a *apiFeature) resetResponse(*godog.Scenario) {
	gin.SetMode(gin.TestMode)
	a.router = gin.New()
	a.response = httptest.NewRecorder()
}

func (a *apiFeature) iHaveAValidGeminiAPIKey() error {
	cfg := config.LoadForTesting()
	if cfg.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY não configurada para testes")
	}
	a.apiKey = cfg.GeminiAPIKey
	a.originalKey = os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", a.apiKey)
	return nil
}

func (a *apiFeature) iDoNotHaveAGeminiAPIKeyConfigured() error {
	a.originalKey = os.Getenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	return nil
}

func (a *apiFeature) iSendAPOSTRequestToRoadmapWithTopic(topic string) error {
	geminiService := services.NewGeminiService(os.Getenv("GEMINI_API_KEY"))
	roadmapHandler := handlers.NewRoadmapHandler(geminiService)
	
	a.router.POST("/roadmap", roadmapHandler.GenerateRoadmap)

	reqBody := map[string]string{"topic": topic}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/roadmap", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	a.router.ServeHTTP(a.response, req)
	return nil
}

func (a *apiFeature) theResponseShouldHaveStatus(status int) error {
	if a.response.Code != status {
		return fmt.Errorf("esperado status %d, mas recebeu %d. Body: %s", status, a.response.Code, a.response.Body.String())
	}
	return nil
}

func (a *apiFeature) theResponseShouldContainARoadmapWithAtLeastCategories(minCategories int) error {
	var roadmap map[string]interface{}
	if err := json.Unmarshal(a.response.Body.Bytes(), &roadmap); err != nil {
		return fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	roadmapArray, ok := roadmap["roadmap"].([]interface{})
	if !ok {
		return fmt.Errorf("roadmap não é um array")
	}

	if len(roadmapArray) < minCategories {
		return fmt.Errorf("esperado pelo menos %d categorias, mas recebeu %d", minCategories, len(roadmapArray))
	}

	return nil
}

func (a *apiFeature) eachCategoryShouldHaveBetweenAndItems(minItems, maxItems int) error {
	var roadmap map[string]interface{}
	if err := json.Unmarshal(a.response.Body.Bytes(), &roadmap); err != nil {
		return fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	roadmapArray, ok := roadmap["roadmap"].([]interface{})
	if !ok {
		return fmt.Errorf("roadmap não é um array")
	}

	for i, cat := range roadmapArray {
		category, ok := cat.(map[string]interface{})
		if !ok {
			return fmt.Errorf("categoria %d não é um objeto", i)
		}

		items, ok := category["items"].([]interface{})
		if !ok {
			return fmt.Errorf("items da categoria %d não é um array", i)
		}

		itemCount := len(items)
		if itemCount < minItems || itemCount > maxItems {
			return fmt.Errorf("categoria %d tem %d itens, mas deveria ter entre %d e %d", i, itemCount, minItems, maxItems)
		}
	}

	return nil
}

func (a *apiFeature) theRoadmapShouldHaveTheTopic(topic string) error {
	var roadmap map[string]interface{}
	if err := json.Unmarshal(a.response.Body.Bytes(), &roadmap); err != nil {
		return fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	roadmapTopic, ok := roadmap["topic"].(string)
	if !ok || roadmapTopic != topic {
		return fmt.Errorf("esperado topic '%s', mas recebeu '%v'", topic, roadmap["topic"])
	}

	return nil
}

func (a *apiFeature) theResponseShouldContainAnErrorMessageAboutAPIKey() error {
	body := a.response.Body.String()
	if body == "" {
		return fmt.Errorf("resposta está vazia")
	}

	var errorResp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &errorResp); err != nil {
		return fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	errorMsg, ok := errorResp["error"].(string)
	if !ok {
		return fmt.Errorf("resposta não contém campo 'error'")
	}

	if errorMsg == "" {
		return fmt.Errorf("mensagem de erro está vazia")
	}

	return nil
}

func InitializeRoadmapScenario(ctx *godog.ScenarioContext) {
	api := &apiFeature{}

	ctx.BeforeScenario(api.resetResponse)

	ctx.Step(`^que tenho uma API key válida do Gemini$`, api.iHaveAValidGeminiAPIKey)
	ctx.Step(`^que não tenho uma API key do Gemini configurada$`, api.iDoNotHaveAGeminiAPIKeyConfigured)
	ctx.Step(`^eu envio uma requisição POST para /roadmap com topic "([^"]*)"$`, api.iSendAPOSTRequestToRoadmapWithTopic)
	ctx.Step(`^a resposta deve ter status (\d+)$`, api.theResponseShouldHaveStatus)
	ctx.Step(`^a resposta deve conter um roadmap com pelo menos (\d+) categorias$`, api.theResponseShouldContainARoadmapWithAtLeastCategories)
	ctx.Step(`^cada categoria deve ter entre (\d+) e (\d+) itens$`, api.eachCategoryShouldHaveBetweenAndItems)
	ctx.Step(`^o roadmap deve ter o topic "([^"]*)"$`, api.theRoadmapShouldHaveTheTopic)
	ctx.Step(`^a resposta deve conter uma mensagem de erro sobre API key$`, api.theResponseShouldContainAnErrorMessageAboutAPIKey)
}

