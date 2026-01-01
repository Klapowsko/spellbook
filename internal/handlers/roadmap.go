package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spellbook/spellbook/internal/models"
	"github.com/spellbook/spellbook/internal/services"
)

// RoadmapHandler gerencia as requisições relacionadas a roadmaps
type RoadmapHandler struct {
	GeminiService services.GeminiServiceInterface
}

// NewRoadmapHandler cria uma nova instância do handler de roadmap
func NewRoadmapHandler(geminiService services.GeminiServiceInterface) *RoadmapHandler {
	return &RoadmapHandler{
		GeminiService: geminiService,
	}
}

// GenerateRoadmap gera um roadmap de estudo
func (h *RoadmapHandler) GenerateRoadmap(c *gin.Context) {
	var req models.RoadmapRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tópico é obrigatório",
		})
		return
	}

	if req.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tópico não pode ser vazio",
		})
		return
	}

	roadmap, err := h.GeminiService.GenerateRoadmap(req.Topic, req.AvailableDays)
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

	c.JSON(http.StatusOK, roadmap)
}

// GenerateEducationalRoadmap gera um roadmap educacional detalhado
func (h *RoadmapHandler) GenerateEducationalRoadmap(c *gin.Context) {
	var req models.EducationalRoadmapRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tópico é obrigatório",
		})
		return
	}

	if req.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tópico não pode ser vazio",
		})
		return
	}

	educationalRoadmap, err := h.GeminiService.GenerateEducationalRoadmap(req.Topic)
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

	c.JSON(http.StatusOK, educationalRoadmap)
}

// GenerateEducationalTrail gera uma trilha educacional estruturada
func (h *RoadmapHandler) GenerateEducationalTrail(c *gin.Context) {
	var req models.EducationalTrailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tópico é obrigatório",
		})
		return
	}

	if req.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tópico não pode ser vazio",
		})
		return
	}

	trail, err := h.GeminiService.GenerateEducationalTrail(req.Topic, req.AvailableDays)
	if err != nil {
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

	c.JSON(http.StatusOK, trail)
}
