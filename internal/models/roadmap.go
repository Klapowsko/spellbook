package models

// RoadmapItem representa um item individual do roadmap
type RoadmapItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// RoadmapCategory representa uma categoria do roadmap
type RoadmapCategory struct {
	Category string        `json:"category"`
	Items    []RoadmapItem `json:"items"`
}

// Roadmap representa o roadmap completo
type Roadmap struct {
	Topic   string            `json:"topic"`
	Roadmap []RoadmapCategory `json:"roadmap"`
}

// RoadmapRequest representa a requisição para gerar um roadmap
type RoadmapRequest struct {
	Topic        string `json:"topic" binding:"required"`
	AvailableDays *int  `json:"available_days,omitempty"`
}

