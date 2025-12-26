package models

// KeyResultsRequest representa a requisição para gerar Key Results
type KeyResultsRequest struct {
	Objective string `json:"objective" binding:"required"`
	Count     int    `json:"count"`
}

// KeyResultsResponse representa a resposta com lista de Key Results
type KeyResultsResponse struct {
	Objective  string   `json:"objective"`
	KeyResults []string `json:"key_results"`
}
