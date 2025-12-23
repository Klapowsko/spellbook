package main

import (
	"log"

	"github.com/spellbook/spellbook/internal/app"
)

func main() {
	// Criar e inicializar aplicação
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Erro ao inicializar aplicação: %v", err)
	}

	// Iniciar servidor
	if err := application.Run(); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
