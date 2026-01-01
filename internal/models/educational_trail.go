package models

// EducationalTrailStep representa uma etapa da trilha educacional
type EducationalTrailStep struct {
	Day         int      `json:"day"`          // Dia da trilha (1, 2, 3...)
	Title       string   `json:"title"`        // Título da etapa (ex: "Dia 1: Fundamentos)
	Description string   `json:"description"` // Descrição do que será feito
	Activities  []Activity `json:"activities"` // Atividades do dia
}

// Activity representa uma atividade específica na trilha
type Activity struct {
	Type        string   `json:"type"`        // "read_book", "read_chapters", "watch_video", "read_article", "do_project", "take_course"
	ResourceID  string   `json:"resource_id"`  // ID do recurso (título do livro, vídeo, etc)
	Title       string   `json:"title"`       // Título da atividade
	Description string   `json:"description"` // Descrição detalhada
	Chapters    []string `json:"chapters,omitempty"` // Capítulos específicos (para livros)
	Duration    string   `json:"duration,omitempty"` // Duração estimada
	URL         string   `json:"url,omitempty"`      // URL do recurso
	Progress    string   `json:"progress,omitempty"` // Progresso esperado (ex: "3 de 5 capítulos")
}

// EducationalTrail representa uma trilha educacional completa
type EducationalTrail struct {
	Topic       string              `json:"topic"`
	TotalDays   int                 `json:"total_days"`
	Description string              `json:"description"`
	Steps       []EducationalTrailStep `json:"steps"`
	Resources   map[string]EducationalResource `json:"resources"` // Recursos referenciados
}

// EducationalTrailRequest representa a requisição para gerar uma trilha educacional
type EducationalTrailRequest struct {
	Topic        string `json:"topic" binding:"required"`
	AvailableDays *int  `json:"available_days,omitempty"`
}

