package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/PavelMkr/docline-new/internal/framework"
)

type OpenAI struct{}

func (a *OpenAI) Name() string {
	return "openai"
}

func (a *OpenAI) Description() string {
	return "Finds fuzzy text duplicates via OpenAI-compatible APIs (OpenAI/Ollama)"
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type llmOutput struct {
	Groups []struct {
		Archetype string   `json:"archetype"`
		Fragments []string `json:"fragments"`
	} `json:"groups"`
}

func (a *OpenAI) FindClones(text string, cfg framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
	provider := strings.ToLower(getString(cfg.CustomParams, "provider", "openai"))
	apiKey := getString(cfg.CustomParams, "openai_api_key", os.Getenv("OPENAI_API_KEY"))
	model := getString(cfg.CustomParams, "model", "gpt-4o-mini")
	userPrompt := getString(cfg.CustomParams, "prompt", "Найди все нечёткие повторы в тексте")
	baseURL := getString(cfg.CustomParams, "base_url", "")
	endpoint := getString(cfg.CustomParams, "endpoint", "")
	timeoutSec := getInt(cfg.CustomParams, "timeout_sec", 90)
	if timeoutSec <= 0 {
		timeoutSec = 90
	}

	resolvedEndpoint, err := resolveEndpoint(provider, baseURL, endpoint)
	if err != nil {
		return nil, err
	}
	requiresAuth := provider != "ollama"
	if requiresAuth && apiKey == "" {
		return nil, fmt.Errorf("openai api key is empty (custom_params.openai_api_key or OPENAI_API_KEY)")
	}

	systemPrompt := `Ты анализатор повторов текста.
	Верни ТОЛЬКО JSON без markdown:
	{"groups":[{"archetype":"...","fragments":["...","..."]}]}
	Правила:
	- группы только с 2+ фрагментами
	- фрагменты должны быть подстроками исходного текста
	- сохраняй оригинальный язык/регистр фрагментов`

	reqPayload := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt + "\n\nТекст:\n" + text},
		},
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	req, err := http.NewRequest(http.MethodPost, resolvedEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create openai request: %w", err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	rawResp, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai status %d: %s", resp.StatusCode, string(rawResp))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(rawResp, &parsed); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	content := cleanupLLMJSON(parsed.Choices[0].Message.Content)

	var out llmOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("decode llm json content: %w; content=%s", err, content)
	}

	groups := make([]framework.CloneGroup, 0, len(out.Groups))
	for _, g := range out.Groups {
		if len(g.Fragments) < 2 {
			continue
		}

		fragments := make([]framework.TextFragment, 0, len(g.Fragments))
		searchFrom := 0
		for _, f := range g.Fragments {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}

			idx := strings.Index(text[searchFrom:], f)
			if idx < 0 {
				idx = strings.Index(text, f)
			} else {
				idx += searchFrom
			}
			if idx < 0 {
				continue
			}

			start := len(strings.Fields(text[:idx]))
			end := start + len(strings.Fields(f))

			fragments = append(fragments, framework.TextFragment{
				Content:  f,
				StartPos: start,
				EndPos:   end,
			})
			searchFrom = idx + len(f)
		}

		if len(fragments) < 2 {
			continue
		}

		sort.Slice(fragments, func(i, j int) bool { return fragments[i].StartPos < fragments[j].StartPos })

		archetype := g.Archetype
		if archetype == "" {
			archetype = fragments[0].Content
		}

		groups = append(groups, framework.CloneGroup{
			Fragments: fragments,
			Power:     len(fragments),
			Archetype: archetype,
			Metadata: map[string]interface{}{
				"finder": "openai-fuzzy",
				"model":  model,
			},
		})
	}

	if cfg.MinGroupPower > 0 {
		filtered := groups[:0]
		for _, g := range groups {
			if len(g.Fragments) >= cfg.MinGroupPower {
				filtered = append(filtered, g)
			}
		}
		groups = filtered
	}

	return groups, nil
}

func resolveEndpoint(provider, baseURL, endpoint string) (string, error) {
	if endpoint != "" {
		return endpoint, nil
	}
	if baseURL == "" {
		if provider == "ollama" {
			baseURL = "http://localhost:11434/v1"
		} else {
			baseURL = "https://api.openai.com/v1"
		}
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return "", fmt.Errorf("invalid base_url %q: expected absolute http(s) URL", baseURL)
	}
	return strings.TrimRight(baseURL, "/") + path.Join("/", "chat", "completions"), nil
}

func cleanupLLMJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
