package models

// TopicsRequest representa a requisição para gerar tópicos
type TopicsRequest struct {
	Subject string `json:"subject" binding:"required"`
	Count   int    `json:"count"`
}

// TopicsResponse representa a resposta com lista de tópicos
type TopicsResponse struct {
	Subject string   `json:"subject"`
	Topics  []string `json:"topics"`
}

