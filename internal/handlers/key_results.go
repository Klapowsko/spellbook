package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spellbook/spellbook/internal/models"
	"github.com/spellbook/spellbook/internal/services"
)

// KeyResultsHandler gerencia as requisições relacionadas a Key Results
type KeyResultsHandler struct {
	GeminiService services.GeminiServiceInterface
}

// NewKeyResultsHandler cria uma nova instância do handler de Key Results
func NewKeyResultsHandler(geminiService services.GeminiServiceInterface) *KeyResultsHandler {
	return &KeyResultsHandler{
		GeminiService: geminiService,
	}
}

// GenerateKeyResults gera uma lista de Key Results mensuráveis para um objetivo OKR
func (h *KeyResultsHandler) GenerateKeyResults(c *gin.Context) {
	var req models.KeyResultsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "objetivo é obrigatório",
		})
		return
	}

	if req.Objective == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "objetivo não pode ser vazio",
		})
		return
	}

	// Se count não foi especificado, usar default de 5
	if req.Count <= 0 {
		req.Count = 5
	}

	keyResults, err := h.GeminiService.GenerateKeyResults(req.Objective, req.Count)
	if err != nil {
		// Verificar se é erro de API key
		if err.Error() == "GEMINI_API_KEY não configurada. Configure no arquivo .env ou variável de ambiente" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "API key do Gemini não configurada",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, keyResults)
}

