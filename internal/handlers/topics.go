package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spellbook/spellbook/internal/models"
	"github.com/spellbook/spellbook/internal/services"
)

// TopicsHandler gerencia as requisições relacionadas a tópicos
type TopicsHandler struct {
	GeminiService services.GeminiServiceInterface
}

// NewTopicsHandler cria uma nova instância do handler de tópicos
func NewTopicsHandler(geminiService services.GeminiServiceInterface) *TopicsHandler {
	return &TopicsHandler{
		GeminiService: geminiService,
	}
}

// GenerateTopics gera uma lista de tópicos sobre um assunto
func (h *TopicsHandler) GenerateTopics(c *gin.Context) {
	var req models.TopicsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "assunto é obrigatório",
		})
		return
	}

	if req.Subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "assunto não pode ser vazio",
		})
		return
	}

	// Se count não foi especificado, usar default de 10
	if req.Count <= 0 {
		req.Count = 10
	}

	topics, err := h.GeminiService.GenerateTopics(req.Subject, req.Count)
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

	c.JSON(http.StatusOK, topics)
}

