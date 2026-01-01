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
			Timeout: 180 * time.Second, // 3 minutos para trilhas educacionais complexas
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
func (s *GeminiService) GenerateRoadmap(topic string, availableDays *int) (*models.Roadmap, error) {
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

	// Determinar número de categorias e itens baseado em availableDays
	numCategories := "4-6"
	itemsPerCategory := "5-10"
	timeContext := ""
	estimatedTotalItems := 30 // Valor padrão
	
	if availableDays != nil && *availableDays > 0 {
		// Calcular proporção de itens baseado no tempo disponível
		// Aproximadamente 1 item por dia, mas permitindo variação natural
		estimatedTotalItems = *availableDays
		
		if *availableDays < 14 {
			// Tempo curto: focar em essencial
			numCategories = "3-4"
			itemsPerCategory = "3-5"
			timeContext = fmt.Sprintf("\n\n⏰ PRAZO CRÍTICO: Este roadmap DEVE ser concluído em EXATAMENTE %d dias.\n\nREGRAS OBRIGATÓRIAS:\n- Crie NO MÁXIMO %d itens no total (não exceda este número)\n- Distribua em 3-4 categorias\n- Cada categoria deve ter entre 3-5 itens\n- Se você criar mais de %d itens, o roadmap será inválido\n- Priorize apenas o ESSENCIAL e mais importante", *availableDays, estimatedTotalItems, estimatedTotalItems)
		} else if *availableDays <= 30 {
			// Tempo médio: estrutura balanceada
			numCategories = "4-6"
			itemsPerCategory = "5-8"
			timeContext = fmt.Sprintf("\n\n⏰ PRAZO CRÍTICO: Este roadmap DEVE ser concluído em EXATAMENTE %d dias.\n\nREGRAS OBRIGATÓRIAS:\n- Crie NO MÁXIMO %d itens no total (não exceda este número)\n- Distribua em 4-6 categorias\n- Cada categoria deve ter entre 5-8 itens\n- Se você criar mais de %d itens, o roadmap será inválido\n- Mantenha uma estrutura balanceada e prática", *availableDays, estimatedTotalItems, estimatedTotalItems)
		} else if *availableDays <= 60 {
			// Tempo médio-longo: estrutura mais completa
			numCategories = "5-7"
			itemsPerCategory = "6-10"
			timeContext = fmt.Sprintf("\n\n⏰ PRAZO CRÍTICO: Este roadmap DEVE ser concluído em EXATAMENTE %d dias.\n\nREGRAS OBRIGATÓRIAS:\n- Crie NO MÁXIMO %d itens no total (não exceda este número)\n- Distribua em 5-7 categorias\n- Cada categoria deve ter entre 6-10 itens\n- Se você criar mais de %d itens, o roadmap será inválido\n- Você tem tempo suficiente para uma estrutura mais completa", *availableDays, estimatedTotalItems, estimatedTotalItems)
		} else {
			// Tempo longo: estrutura extensa mas organizada
			// Calcular estimativas proporcionais ao tempo
			estimatedCategories := 6 + (*availableDays-60)/15 // Aproximadamente 1 categoria a cada 15 dias extras
			itemsPerCat := estimatedTotalItems / estimatedCategories
			
			numCategories = fmt.Sprintf("%d-%d", estimatedCategories-1, estimatedCategories+2)
			itemsPerCategory = fmt.Sprintf("%d-%d", itemsPerCat-2, itemsPerCat+3)
			timeContext = fmt.Sprintf("\n\n⏰ PRAZO CRÍTICO: Este roadmap DEVE ser concluído em EXATAMENTE %d dias.\n\nREGRAS OBRIGATÓRIAS:\n- Crie NO MÁXIMO %d itens no total (não exceda este número)\n- Distribua em %s categorias\n- Cada categoria deve ter entre %s itens\n- Se você criar mais de %d itens, o roadmap será inválido\n- A quantidade deve ser proporcional ao tempo disponível", *availableDays, estimatedTotalItems, numCategories, itemsPerCategory, estimatedTotalItems)
		}
	}

	// Prompt para gerar o roadmap
	prompt := fmt.Sprintf(`Você é um especialista em criar roadmaps de estudo detalhados e estruturados.

Crie um roadmap completo e bem organizado sobre: "%s"%s

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

Requisitos OBRIGATÓRIOS:
- Crie EXATAMENTE %s categorias principais (não mais, não menos)
- Cada categoria deve ter EXATAMENTE entre %s itens (respeite este intervalo)
- O TOTAL DE ITENS em todo o roadmap NÃO DEVE EXCEDER %d itens
- Os itens devem ser progressivos (do básico ao avançado)
- Seja específico e prático nos títulos
- Organize de forma lógica e sequencial

VALIDAÇÃO: Se o roadmap tiver mais de %d itens totais, ele será rejeitado e você terá que gerar novamente.

Retorne APENAS o JSON válido, sem markdown code blocks, sem texto antes ou depois.`, topic, timeContext, topic, numCategories, itemsPerCategory, estimatedTotalItems, estimatedTotalItems)

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

		// Validar quantidade total de itens
		totalItems := 0
		for _, category := range roadmap.Roadmap {
			totalItems += len(category.Items)
		}

		// Se availableDays foi fornecido, validar se o número de itens está dentro do esperado
		if availableDays != nil && *availableDays > 0 {
			maxExpectedItems := *availableDays + 5 // Permitir 5 itens a mais como margem
			if totalItems > maxExpectedItems {
				lastError = fmt.Errorf("roadmap gerado com %d itens, mas o limite é %d itens (tempo disponível: %d dias). Tentando novamente...", totalItems, maxExpectedItems, *availableDays)
				// Tentar novamente com o mesmo modelo, mas com prompt mais restritivo
				continue
			}
			// Log para debug
			fmt.Printf("[DEBUG] Spellbook GenerateRoadmap - AvailableDays: %d, TotalItemsGenerated: %d, MaxExpected: %d\n", 
				*availableDays, totalItems, maxExpectedItems)
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

// getTimeDistributionInstructions retorna instruções sobre como distribuir Key Results no tempo
func getTimeDistributionInstructions(completionDate *string) string {
	if completionDate == nil || *completionDate == "" {
		return ""
	}

	completionTime, err := time.Parse("2006-01-02", *completionDate)
	if err != nil {
		return ""
	}

	now := time.Now()
	daysRemaining := int(completionTime.Sub(now).Hours() / 24)
	monthsRemaining := daysRemaining / 30

	if daysRemaining < 0 {
		return "Distribuição temporal: Todos os Key Results devem ser realizáveis imediatamente, priorizando resultados rápidos."
	} else if monthsRemaining < 3 {
		return "Distribuição temporal: Todos os Key Results devem ser realizáveis em curto prazo (semanas). Priorize resultados rápidos e simples."
	} else if monthsRemaining <= 6 {
		return fmt.Sprintf("Distribuição temporal: Distribua os Key Results ao longo de %d meses - alguns no primeiro mês (início), outros no meio do período, e alguns no final. Complexidade moderada.", monthsRemaining)
	} else {
		return fmt.Sprintf("Distribuição temporal: Distribua os Key Results progressivamente ao longo de %d meses - Key Results iniciais (primeiro mês), intermediários (meio do período), e finais (último mês). Pode incluir Key Results mais complexos e ambiciosos.", monthsRemaining)
	}
}

// GenerateKeyResults gera uma lista de Key Results mensuráveis para um objetivo OKR
func (s *GeminiService) GenerateKeyResults(objective string, count int, completionDate *string) (*models.KeyResultsResponse, error) {
	if objective == "" {
		return nil, fmt.Errorf("objetivo não pode ser vazio")
	}

	if count <= 0 {
		count = 5 // Default para Key Results
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

	// Calcular informações sobre o prazo
	var timeContext string
	if completionDate != nil && *completionDate != "" {
		completionTime, err := time.Parse("2006-01-02", *completionDate)
		if err == nil {
			now := time.Now()
			daysRemaining := int(completionTime.Sub(now).Hours() / 24)
			monthsRemaining := daysRemaining / 30
			
			if daysRemaining < 0 {
				timeContext = fmt.Sprintf("\n\n⚠️ ATENÇÃO: A data de conclusão (%s) já passou. Ajuste os Key Results para serem realizáveis no menor tempo possível.", *completionDate)
			} else if monthsRemaining < 3 {
				timeContext = fmt.Sprintf("\n\n⏰ PRAZO: Este OKR deve ser concluído em %d dias (menos de 3 meses). Gere Key Results SIMPLES, DIRETOS e REALIZÁVEIS no curto prazo. Priorize resultados rápidos e de baixa complexidade.", daysRemaining)
			} else if monthsRemaining <= 6 {
				timeContext = fmt.Sprintf("\n\n⏰ PRAZO: Este OKR deve ser concluído em %d dias (aproximadamente %d meses). Distribua os Key Results ao longo do tempo: alguns no primeiro mês, outros no meio do período, e alguns no final. Complexidade MODERADA.", daysRemaining, monthsRemaining)
			} else {
				timeContext = fmt.Sprintf("\n\n⏰ PRAZO: Este OKR deve ser concluído em %d dias (aproximadamente %d meses). Distribua os Key Results progressivamente: Key Results iniciais (primeiro mês), intermediários (meio do período), e finais (último mês). Pode incluir Key Results mais complexos e ambiciosos.", daysRemaining, monthsRemaining)
			}
		}
	}

	// Prompt específico para gerar Key Results mensuráveis para OKRs
	prompt := fmt.Sprintf(`Você é um especialista em OKRs (Objectives and Key Results).

Gere uma lista de %d Key Results mensuráveis e específicos para o seguinte objetivo: "%s"%s

Key Results devem ser:
- Mensuráveis (com métricas claras)
- Específicos e acionáveis
- Alinhados com o objetivo
- Focados em resultados, não apenas em atividades
- Realistas e alcançáveis

%s

A resposta deve ser APENAS um JSON válido, sem markdown, sem texto adicional, seguindo EXATAMENTE esta estrutura:

{
  "objective": "%s",
  "key_results": [
    "Key Result 1",
    "Key Result 2",
    "Key Result 3"
  ]
}

Requisitos:
- Cada Key Result deve ser uma frase clara e mensurável
- Use métricas específicas quando possível (números, percentuais, etc.)
- Foque em resultados que demonstrem progresso em direção ao objetivo
- Seja conciso mas específico
- Retorne APENAS o JSON, sem explicações adicionais

IMPORTANTE: Retorne apenas o JSON válido, sem markdown code blocks, sem texto antes ou depois.`, count, objective, timeContext, getTimeDistributionInstructions(completionDate), objective)

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

		var keyResultsResp models.KeyResultsResponse
		if err := json.Unmarshal([]byte(jsonText), &keyResultsResp); err != nil {
			lastError = fmt.Errorf("erro ao fazer parse do JSON: %v", err)
			continue
		}

		if keyResultsResp.Objective == "" || len(keyResultsResp.KeyResults) == 0 {
			lastError = fmt.Errorf("resposta do Gemini não está no formato esperado")
			continue
		}

		return &keyResultsResp, nil
	}

	if lastError != nil {
		return nil, fmt.Errorf("erro ao gerar Key Results: %v", lastError)
	}

	return nil, fmt.Errorf("erro ao gerar Key Results: nenhum modelo disponível funcionou")
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
func (s *GeminiService) GenerateEducationalTrail(topic string, availableDays *int) (*models.EducationalTrail, error) {
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

	// Determinar dias totais e atividades por dia baseado em availableDays
	totalDays := 12
	activitiesPerDay := "2-3"
	timeContext := ""
	
	if availableDays != nil && *availableDays > 0 {
		totalDays = *availableDays
		
		if totalDays < 7 {
			// Tempo curto: focar em essencial, menos atividades
			activitiesPerDay = "1-2"
			timeContext = fmt.Sprintf("\n\n⏰ PRAZO LIMITADO: Esta trilha deve ser concluída em %d dias. Foque em conteúdo ESSENCIAL e DIRETO. Priorize atividades rápidas e práticas. Menos atividades por dia (1-2), mas bem focadas.", totalDays)
		} else if totalDays <= 14 {
			// Tempo médio: estrutura balanceada
			activitiesPerDay = "2-3"
			timeContext = fmt.Sprintf("\n\n⏰ PRAZO: Esta trilha deve ser concluída em %d dias. Mantenha um ritmo balanceado com 2-3 atividades por dia.", totalDays)
		} else {
			// Tempo longo: conteúdo mais aprofundado
			activitiesPerDay = "3-4"
			timeContext = fmt.Sprintf("\n\n⏰ PRAZO: Esta trilha deve ser concluída em %d dias. Você tem tempo suficiente para conteúdo mais aprofundado. Pode incluir 3-4 atividades por dia e materiais mais extensos.", totalDays)
		}
	}

	// Prompt para gerar trilha educacional estruturada (otimizado para ser mais rápido)
	prompt := fmt.Sprintf(`Crie uma trilha educacional de %d dias sobre: "%s"%s

Retorne APENAS JSON válido, sem markdown:

{
  "topic": "%s",
  "total_days": %d,
  "description": "Trilha de aprendizado progressiva",
  "resources": {
    "recurso_1": {"title": "Nome", "description": "Desc", "author": "Autor", "chapters": ["Cap 1"], "url": ""},
    "recurso_2": {"title": "Vídeo", "duration": "30 min", "url": ""}
  },
  "steps": [
    {
      "day": 1,
      "title": "Dia 1: Título",
      "description": "O que será aprendido",
      "activities": [
        {
          "type": "read_chapters",
          "resource_id": "recurso_1",
          "title": "Ler capítulos 1-3",
          "description": "Foque em...",
          "chapters": ["Cap 1", "Cap 2"],
          "progress": "3 de 10 capítulos"
        }
      ]
    }
  ]
}

Regras IMPORTANTES:
- EXATAMENTE %d dias, %s atividades por dia
- O campo "total_days" no JSON DEVE ser %d
- Tipos: read_chapters, watch_video, read_article, take_course, do_project
- Progressivo: básico → avançado → prática
- Seja específico: "Ler capítulos 1-3" não "Ler livro"
- Inclua progresso quando relevante
- Projetos no final
- Distribua o conteúdo proporcionalmente ao longo dos %d dias

CRITÉRIOS PARA RECURSOS (LIVROS, CURSOS, VÍDEOS, ARTIGOS):
- Use APENAS recursos amplamente conhecidos, estabelecidos e reconhecidos na área
- Priorize recursos clássicos, best-sellers e materiais amplamente utilizados
- Evite recursos muito recentes, específicos ou obscuros que podem não existir
- Para livros: use apenas livros famosos, best-sellers ou clássicos da área (ex: "Clean Code", "Design Patterns", "The Pragmatic Programmer")
- Para cursos: use plataformas conhecidas (Coursera, edX, Udemy) e cursos populares/verificados
- Para vídeos: use canais conhecidos e vídeos populares (YouTube, com muitos views)
- Para artigos: use artigos de sites conhecidos e estabelecidos
- Se não tiver certeza se um recurso existe, prefira recursos genéricos ou bem conhecidos
- URLs devem ser válidas e acessíveis - evite URLs quebradas ou inexistentes
- Se não souber uma URL específica, deixe o campo "url" vazio ao invés de inventar uma

APENAS JSON, sem markdown.`, totalDays, topic, timeContext, topic, totalDays, activitiesPerDay, totalDays, totalDays)

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
