package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config armazena as configurações da aplicação
type Config struct {
	GeminiAPIKey string
	Port         string
}

// Load carrega as configurações do ambiente
func Load() (*Config, error) {
	// Tentar carregar .env (não é erro se não existir)
	_ = godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY não configurada. Configure no arquivo .env ou variável de ambiente")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		GeminiAPIKey: apiKey,
		Port:         port,
	}, nil
}

// LoadForTesting carrega configurações para testes (permite API key vazia)
func LoadForTesting() *Config {
	_ = godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		GeminiAPIKey: apiKey,
		Port:         port,
	}
}
