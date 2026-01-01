package models

// KeyResultsRequest representa a requisição para gerar Key Results
type KeyResultsRequest struct {
	Objective      string  `json:"objective" binding:"required"`
	Count          int     `json:"count"`
	CompletionDate *string `json:"completion_date,omitempty"`
}

// KeyResultsResponse representa a resposta com lista de Key Results
type KeyResultsResponse struct {
	Objective  string   `json:"objective"`
	KeyResults []string `json:"key_results"`
}
