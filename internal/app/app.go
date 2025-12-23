package app

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spellbook/spellbook/internal/config"
	"github.com/spellbook/spellbook/internal/handlers"
	"github.com/spellbook/spellbook/internal/routes"
	"github.com/spellbook/spellbook/internal/services"
)

// App representa a aplicação e suas dependências
type App struct {
	Config         *config.Config
	GeminiService  *services.GeminiService
	RoadmapHandler *handlers.RoadmapHandler
	TopicsHandler  *handlers.TopicsHandler
	Router         *gin.Engine
}

// NewApp cria e inicializa uma nova instância da aplicação
func NewApp() (*App, error) {
	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar configurações: %w", err)
	}

	// Criar serviço Gemini
	geminiService := services.NewGeminiService(cfg.GeminiAPIKey)

	// Criar handlers
	roadmapHandler := handlers.NewRoadmapHandler(geminiService)
	topicsHandler := handlers.NewTopicsHandler(geminiService)

	// Configurar Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Configurar rotas
	routes.SetupRoutes(router, roadmapHandler, topicsHandler)

	return &App{
		Config:         cfg,
		GeminiService:  geminiService,
		RoadmapHandler: roadmapHandler,
		TopicsHandler:  topicsHandler,
		Router:         router,
	}, nil
}

// Run inicia o servidor HTTP
func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.Config.Port)
	log.Printf("Servidor Spellbook iniciado na porta %s", a.Config.Port)
	log.Printf("Health check: http://localhost%s/health", addr)
	log.Printf("API disponível em: http://localhost%s/roadmap e http://localhost%s/topics", addr, addr)

	return a.Router.Run(addr)
}

