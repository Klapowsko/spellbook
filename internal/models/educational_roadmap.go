package models

// EducationalResource representa um recurso educacional (livro, curso, vídeo, etc)
type EducationalResource struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url,omitempty"`
	Chapters    []string `json:"chapters,omitempty"`
	Duration    string   `json:"duration,omitempty"`
	Author      string   `json:"author,omitempty"`
}

// EducationalRoadmap representa um roadmap educacional completo
type EducationalRoadmap struct {
	Topic    string                `json:"topic"`
	Books    []EducationalResource `json:"books"`
	Courses  []EducationalResource `json:"courses"`
	Videos   []EducationalResource `json:"videos"`
	Articles []EducationalResource `json:"articles"`
	Projects []EducationalResource `json:"projects"`
}

// EducationalRoadmapRequest representa a requisição para gerar um roadmap educacional
type EducationalRoadmapRequest struct {
	Topic string `json:"topic" binding:"required"`
}

