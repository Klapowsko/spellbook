package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/spellbook/spellbook/internal/models"
)

// GeminiService gerencia a integração com a API do Gemini
type GeminiService struct {
	APIKey     string
	HTTPClient *http.Client
	BaseURL    string
}

// NewGeminiService cria uma nova instância do serviço Gemini
func NewGeminiService(apiKey string) *GeminiService {
	return &GeminiService{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		BaseURL: "https://generativelanguage.googleapis.com/v1beta",
	}
}

// listAvailableModels lista os modelos disponíveis na API
func (s *GeminiService) listAvailableModels() ([]string, error) {
	url := fmt.Sprintf("%s/models?key=%s", s.BaseURL, s.APIKey)

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, nil
	}

	var data struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return []string{}, nil
	}

	models := make([]string, 0)
	for _, model := range data.Models {
		if model.Name != "" {
			name := strings.TrimPrefix(model.Name, "models/")
			if strings.Contains(name, "gemini") && !strings.Contains(name, "embedding") {
				models = append(models, name)
			}
		}
	}

	return models, nil
}

// generateContent gera conteúdo usando um modelo específico
func (s *GeminiService) generateContent(modelName, prompt string) (string, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", s.BaseURL, modelName, s.APIKey)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		// Erro de quota - retornar erro especial
		return "", fmt.Errorf("quota excedida (429)")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro da API: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// cleanJSONText limpa o texto para extrair apenas o JSON
func cleanJSONText(text string) string {
	// Remover markdown code blocks
	text = regexp.MustCompile("(?i)```json\\s*").ReplaceAllString(text, "")
	text = regexp.MustCompile("(?i)```\\s*").ReplaceAllString(text, "")
	text = strings.TrimSpace(text)

	// Tentar encontrar o JSON no texto
	jsonMatch := regexp.MustCompile(`\{[\s\S]*\}`).FindString(text)
	if jsonMatch != "" {
		return jsonMatch
	}

	return text
}

// GenerateRoadmap gera um roadmap de estudo usando o Gemini
func (s *GeminiService) GenerateRoadmap(topic string) (*models.Roadmap, error) {
	if topic == "" {
		return nil, fmt.Errorf("tópico não pode ser vazio")
	}

	// Listar modelos disponíveis
	availableModels, _ := s.listAvailableModels()

	// Lista de modelos para tentar em ordem (fallback)
	modelsToTry := make([]string, 0)

	// Adicionar modelos disponíveis primeiro
	for _, model := range availableModels {
		modelsToTry = append(modelsToTry, model)
	}

	// Adicionar fallbacks
	fallbacks := []string{
		"gemini-1.5-flash-latest",
		"gemini-1.5-pro-latest",
		"gemini-pro",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
	}

	// Remover duplicatas
	seen := make(map[string]bool)
	for _, model := range modelsToTry {
		seen[model] = true
	}
	for _, model := range fallbacks {
		if !seen[model] {
			modelsToTry = append(modelsToTry, model)
		}
	}

	// Prompt para gerar o roadmap
	prompt := fmt.Sprintf(`Você é um especialista em criar roadmaps de estudo detalhados e estruturados.

Crie um roadmap completo e bem organizado sobre: "%s"

O roadmap deve ser retornado APENAS como um JSON válido, sem markdown, sem texto adicional, seguindo EXATAMENTE esta estrutura:

{
  "topic": "%s",
  "roadmap": [
    {
      "category": "Nome da Categoria",
      "items": [
        {"id": "1", "title": "Título do item", "completed": false},
        {"id": "2", "title": "Título do item", "completed": false}
      ]
    }
  ]
}

Requisitos:
- Crie pelo menos 4-6 categorias principais
- Cada categoria deve ter entre 5-10 itens
- Os itens devem ser progressivos (do básico ao avançado)
- Seja específico e prático nos títulos
- Organize de forma lógica e sequencial
- Retorne APENAS o JSON, sem explicações adicionais

IMPORTANTE: Retorne apenas o JSON válido, sem markdown code blocks, sem texto antes ou depois.`, topic, topic)

	var lastError error

	// Tentar cada modelo até encontrar um que funcione
	for _, modelName := range modelsToTry {
		text, err := s.generateContent(modelName, prompt)
		if err != nil {
			// Se for erro de quota, aguardar e tentar novamente
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
				time.Sleep(30 * time.Second)
				// Tentar novamente este modelo
				text, err = s.generateContent(modelName, prompt)
				if err != nil {
					lastError = err
					continue
				}
			} else {
				lastError = err
				continue
			}
		}

		// Limpar o texto para extrair apenas o JSON
		jsonText := cleanJSONText(text)

		// Tentar fazer parse do JSON
		var roadmap models.Roadmap
		if err := json.Unmarshal([]byte(jsonText), &roadmap); err != nil {
			lastError = fmt.Errorf("erro ao fazer parse do JSON: %v", err)
			continue
		}

		// Validar estrutura básica
		if roadmap.Topic == "" || len(roadmap.Roadmap) == 0 {
			lastError = fmt.Errorf("resposta do Gemini não está no formato esperado")
			continue
		}

		return &roadmap, nil
	}

	if lastError != nil {
		return nil, fmt.Errorf("erro ao gerar roadmap: %v", lastError)
	}

	return nil, fmt.Errorf("erro ao gerar roadmap: nenhum modelo disponível funcionou")
}

// GenerateTopics gera uma lista de tópicos sobre um assunto
func (s *GeminiService) GenerateTopics(subject string, count int) (*models.TopicsResponse, error) {
	if subject == "" {
		return nil, fmt.Errorf("assunto não pode ser vazio")
	}

	if count <= 0 {
		count = 10 // Default
	}

	// Listar modelos disponíveis
	availableModels, _ := s.listAvailableModels()

	modelsToTry := make([]string, 0)
	seen := make(map[string]bool)

	for _, model := range availableModels {
		modelsToTry = append(modelsToTry, model)
		seen[model] = true
	}

	fallbacks := []string{
		"gemini-1.5-flash-latest",
		"gemini-1.5-pro-latest",
		"gemini-pro",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
	}

	for _, model := range fallbacks {
		if !seen[model] {
			modelsToTry = append(modelsToTry, model)
		}
	}

	// Prompt para gerar tópicos
	prompt := fmt.Sprintf(`Você é um especialista em organizar conhecimento.

Gere uma lista de %d tópicos importantes e relevantes sobre: "%s"

A resposta deve ser APENAS um JSON válido, sem markdown, sem texto adicional, seguindo EXATAMENTE esta estrutura:

{
  "subject": "%s",
  "topics": [
    "Tópico 1",
    "Tópico 2",
    "Tópico 3"
  ]
}

Requisitos:
- Liste tópicos práticos e específicos
- Organize de forma lógica
- Seja conciso nos nomes dos tópicos
- Retorne APENAS o JSON, sem explicações adicionais

IMPORTANTE: Retorne apenas o JSON válido, sem markdown code blocks, sem texto antes ou depois.`, count, subject, subject)

	var lastError error

	for _, modelName := range modelsToTry {
		text, err := s.generateContent(modelName, prompt)
		if err != nil {
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
				time.Sleep(30 * time.Second)
				text, err = s.generateContent(modelName, prompt)
				if err != nil {
					lastError = err
					continue
				}
			} else {
				lastError = err
				continue
			}
		}

		jsonText := cleanJSONText(text)

		var topicsResp models.TopicsResponse
		if err := json.Unmarshal([]byte(jsonText), &topicsResp); err != nil {
			lastError = fmt.Errorf("erro ao fazer parse do JSON: %v", err)
			continue
		}

		if topicsResp.Subject == "" || len(topicsResp.Topics) == 0 {
			lastError = fmt.Errorf("resposta do Gemini não está no formato esperado")
			continue
		}

		return &topicsResp, nil
	}

	if lastError != nil {
		return nil, fmt.Errorf("erro ao gerar tópicos: %v", lastError)
	}

	return nil, fmt.Errorf("erro ao gerar tópicos: nenhum modelo disponível funcionou")
}

// GenerateEducationalRoadmap gera um roadmap educacional detalhado com livros, cursos, vídeos, artigos e projetos
func (s *GeminiService) GenerateEducationalRoadmap(topic string) (*models.EducationalRoadmap, error) {
	if topic == "" {
		return nil, fmt.Errorf("tópico não pode ser vazio")
	}

	// Listar modelos disponíveis
	availableModels, _ := s.listAvailableModels()

	modelsToTry := make([]string, 0)
	seen := make(map[string]bool)

	for _, model := range availableModels {
		modelsToTry = append(modelsToTry, model)
		seen[model] = true
	}

	fallbacks := []string{
		"gemini-1.5-flash-latest",
		"gemini-1.5-pro-latest",
		"gemini-pro",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
	}

	for _, model := range fallbacks {
		if !seen[model] {
			modelsToTry = append(modelsToTry, model)
		}
	}

	// Prompt para gerar roadmap educacional
	prompt := fmt.Sprintf(`Você é um especialista em criar roadmaps educacionais detalhados e estruturados.

Crie um roadmap educacional completo e bem organizado sobre: "%s"

O roadmap deve ser retornado APENAS como um JSON válido, sem markdown, sem texto adicional, seguindo EXATAMENTE esta estrutura:

{
  "topic": "%s",
  "books": [
    {
      "title": "Nome do Livro",
      "description": "Descrição do livro",
      "author": "Nome do Autor",
      "chapters": ["Capítulo 1", "Capítulo 2", "Capítulo 3"],
      "url": "URL do livro (se disponível)"
    }
  ],
  "courses": [
    {
      "title": "Nome do Curso",
      "description": "Descrição do curso",
      "duration": "Duração estimada",
      "url": "URL do curso"
    }
  ],
  "videos": [
    {
      "title": "Nome do Vídeo",
      "description": "Descrição do vídeo",
      "duration": "Duração do vídeo",
      "url": "URL do vídeo"
    }
  ],
  "articles": [
    {
      "title": "Nome do Artigo",
      "description": "Descrição do artigo",
      "url": "URL do artigo"
    }
  ],
  "projects": [
    {
      "title": "Nome do Projeto",
      "description": "Descrição do projeto lúdico para consolidar conhecimento",
      "url": "URL de referência (se disponível)"
    }
  ]
}

Requisitos:
- Inclua 3-5 livros relevantes com seus principais capítulos
- Inclua 3-5 cursos online ou presenciais
- Inclua 5-10 vídeos educacionais (YouTube, etc)
- Inclua 5-10 artigos técnicos ou tutoriais
- Inclua 3-5 projetos práticos e lúdicos para consolidar o conhecimento
- Seja específico e prático nas descrições
- Organize de forma progressiva (do básico ao avançado)
- Retorne APENAS o JSON, sem explicações adicionais

IMPORTANTE: Retorne apenas o JSON válido, sem markdown code blocks, sem texto antes ou depois.`, topic, topic)

	var lastError error

	for _, modelName := range modelsToTry {
		text, err := s.generateContent(modelName, prompt)
		if err != nil {
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
				time.Sleep(30 * time.Second)
				text, err = s.generateContent(modelName, prompt)
				if err != nil {
					lastError = err
					continue
				}
			} else {
				lastError = err
				continue
			}
		}

		jsonText := cleanJSONText(text)

		var educationalRoadmap models.EducationalRoadmap
		if err := json.Unmarshal([]byte(jsonText), &educationalRoadmap); err != nil {
			lastError = fmt.Errorf("erro ao fazer parse do JSON: %v", err)
			continue
		}

		// Validar estrutura básica
		if educationalRoadmap.Topic == "" {
			lastError = fmt.Errorf("resposta do Gemini não está no formato esperado")
			continue
		}

		return &educationalRoadmap, nil
	}

	if lastError != nil {
		return nil, fmt.Errorf("erro ao gerar roadmap educacional: %v", lastError)
	}

	return nil, fmt.Errorf("erro ao gerar roadmap educacional: nenhum modelo disponível funcionou")
}

// GenerateEducationalTrail gera uma trilha educacional estruturada em dias/etapas
func (s *GeminiService) GenerateEducationalTrail(topic string) (*models.EducationalTrail, error) {
	if topic == "" {
		return nil, fmt.Errorf("tópico não pode ser vazio")
	}

	// Listar modelos disponíveis
	availableModels, _ := s.listAvailableModels()

	modelsToTry := make([]string, 0)
	seen := make(map[string]bool)

	for _, model := range availableModels {
		modelsToTry = append(modelsToTry, model)
		seen[model] = true
	}

	fallbacks := []string{
		"gemini-1.5-flash-latest",
		"gemini-1.5-pro-latest",
		"gemini-pro",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
	}

	for _, model := range fallbacks {
		if !seen[model] {
			modelsToTry = append(modelsToTry, model)
		}
	}

	// Prompt para gerar trilha educacional estruturada
	prompt := fmt.Sprintf(`Você é um especialista em criar trilhas de aprendizado estruturadas e progressivas.

Crie uma trilha educacional completa e bem organizada sobre: "%s"

A trilha deve ser organizada em DIAS/ETAPAS, onde cada dia tem atividades específicas e progressivas. 
Organize os recursos (livros, cursos, vídeos, artigos, projetos) em uma sequência lógica de aprendizado.

O roadmap deve ser retornado APENAS como um JSON válido, sem markdown, sem texto adicional, seguindo EXATAMENTE esta estrutura:

{
  "topic": "%s",
  "total_days": 14,
  "description": "Descrição geral da trilha",
  "resources": {
    "livro_clean_code": {
      "title": "Clean Code",
      "description": "Descrição",
      "author": "Robert C. Martin",
      "chapters": ["Capítulo 1", "Capítulo 2", ...],
      "url": "URL (opcional)"
    },
    "video_solid_principles": {
      "title": "SOLID Principles Explained",
      "description": "Descrição",
      "duration": "30 min",
      "url": "URL"
    }
  },
  "steps": [
    {
      "day": 1,
      "title": "Dia 1: Fundamentos e Introdução",
      "description": "Neste dia você vai aprender os conceitos básicos...",
      "activities": [
        {
          "type": "read_chapters",
          "resource_id": "livro_clean_code",
          "title": "Ler capítulos 1-3 do livro Clean Code",
          "description": "Foque em entender os princípios de código limpo",
          "chapters": ["Capítulo 1: Código Limpo", "Capítulo 2: Nomes Significativos", "Capítulo 3: Funções"],
          "progress": "3 de 17 capítulos"
        },
        {
          "type": "watch_video",
          "resource_id": "video_solid_principles",
          "title": "Assistir vídeo sobre SOLID",
          "description": "Entenda os 5 princípios SOLID",
          "duration": "30 min",
          "url": "URL do vídeo"
        },
        {
          "type": "read_article",
          "resource_id": "artigo_refactoring",
          "title": "Ler artigo sobre Refactoring",
          "description": "Aprenda técnicas de refatoração",
          "url": "URL do artigo"
        }
      ]
    },
    {
      "day": 2,
      "title": "Dia 2: Aprofundamento",
      "description": "Continue aprendendo...",
      "activities": [...]
    }
  ]
}

Tipos de atividades disponíveis:
- "read_book": Ler um livro completo
- "read_chapters": Ler capítulos específicos de um livro
- "watch_video": Assistir um vídeo
- "read_article": Ler um artigo
- "take_course": Fazer um curso (pode ser dividido em partes)
- "do_project": Fazer um projeto prático

Requisitos IMPORTANTES:
- Crie uma trilha de 10-21 dias (dependendo da complexidade do tópico)
- Cada dia deve ter 2-4 atividades bem definidas
- Organize de forma progressiva: do básico ao avançado
- Distribua os recursos ao longo dos dias de forma equilibrada
- Para livros, divida em capítulos ao longo de vários dias
- Para cursos, divida em módulos/aulas
- Projetos devem vir no final ou distribuídos conforme o aprendizado
- Seja específico: "Ler capítulos 1-3" ao invés de "Ler livro"
- Inclua progresso: "3 de 17 capítulos", "50%% do curso", etc
- Cada atividade deve ter uma descrição clara do que fazer
- Use resource_id para referenciar recursos no objeto "resources"

Exemplo de distribuição:
- Dias 1-5: Fundamentos (leituras iniciais, vídeos introdutórios)
- Dias 6-10: Aprofundamento (mais capítulos, artigos técnicos)
- Dias 11-15: Prática (projetos, exercícios)
- Dias 16+: Consolidação (revisão, projetos finais)

Retorne APENAS o JSON válido, sem markdown code blocks, sem texto antes ou depois.`, topic, topic)

	var lastError error

	for _, modelName := range modelsToTry {
		text, err := s.generateContent(modelName, prompt)
		if err != nil {
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
				time.Sleep(30 * time.Second)
				text, err = s.generateContent(modelName, prompt)
				if err != nil {
					lastError = err
					continue
				}
			} else {
				lastError = err
				continue
			}
		}

		jsonText := cleanJSONText(text)

		var trail models.EducationalTrail
		if err := json.Unmarshal([]byte(jsonText), &trail); err != nil {
			lastError = fmt.Errorf("erro ao fazer parse do JSON: %v", err)
			continue
		}

		// Validar estrutura básica
		if trail.Topic == "" || len(trail.Steps) == 0 {
			lastError = fmt.Errorf("resposta do Gemini não está no formato esperado")
			continue
		}

		return &trail, nil
	}

	if lastError != nil {
		return nil, fmt.Errorf("erro ao gerar trilha educacional: %v", lastError)
	}

	return nil, fmt.Errorf("erro ao gerar trilha educacional: nenhum modelo disponível funcionou")
}
