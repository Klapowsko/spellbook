package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/spellbook/spellbook/internal/handlers"
	"github.com/spellbook/spellbook/internal/middleware"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(router *gin.Engine, roadmapHandler *handlers.RoadmapHandler, topicsHandler *handlers.TopicsHandler) {
	// Aplicar middleware global
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "spellbook",
		})
	})

	// Rotas da API com prefixo /api/v1
	api := router.Group("/api/v1")
	{
		api.POST("/roadmap", roadmapHandler.GenerateRoadmap)
		api.POST("/topics", topicsHandler.GenerateTopics)
		api.POST("/educational-roadmap", roadmapHandler.GenerateEducationalRoadmap)
	}

	// Rotas sem prefixo /api/v1 (para compatibilidade)
	router.POST("/roadmap", roadmapHandler.GenerateRoadmap)
	router.POST("/topics", topicsHandler.GenerateTopics)
	router.POST("/educational-roadmap", roadmapHandler.GenerateEducationalRoadmap)
}

