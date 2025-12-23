package features

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/spellbook/spellbook/features/step_definitions"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			step_definitions.InitializeRoadmapScenario(ctx)
			step_definitions.InitializeTopicsScenario(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestMain(m *testing.M) {
	// Configurar variáveis de ambiente para testes se necessário
	// Isso permite que os testes funcionem mesmo sem .env
	_ = os.Setenv("GEMINI_API_KEY", os.Getenv("GEMINI_API_KEY"))

	status := m.Run()
	os.Exit(status)
}

