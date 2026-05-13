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
	"unicode/utf8"

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
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
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

	chunkMaxRunes := getInt(cfg.CustomParams, "openai_chunk_max_runes", 48000)
	if chunkMaxRunes < 2048 {
		chunkMaxRunes = 2048
	}
	overlapRunes := getInt(cfg.CustomParams, "openai_chunk_overlap_runes", 2000)
	if overlapRunes < 0 {
		overlapRunes = 0
	}
	if overlapRunes >= chunkMaxRunes {
		overlapRunes = chunkMaxRunes / 8
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

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	runes := []rune(text)
	if len(runes) == 0 {
		return nil, nil
	}

	ensureCustomParams(&cfg)

	var allGroups []framework.CloneGroup
	var totalLLMGroups, totalDroppedFrags int
	var sumPrompt, sumCompletion, sumTotal int

	chunks := chunkByteSpans(text, chunkMaxRunes, overlapRunes)
	if len(chunks) > 1 {
		fmt.Fprintf(os.Stderr, "openai: the text is large (%d runes), %d queries are performed on ~%d runes with %d overlap (see custom_params openai_chunk_*)\n",
			len(runes), len(chunks), chunkMaxRunes, overlapRunes)
	}

	for ci, span := range chunks {
		chunk := text[span[0]:span[1]]
		reqPayload := openAIRequest{
			Model: model,
			Messages: []openAIMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt + "\n\nТекст:\n" + chunk},
			},
		}

		content, usage, err := openaiChatOnce(client, resolvedEndpoint, apiKey, reqPayload)
		if err != nil {
			return nil, err
		}
		sumPrompt += usage.PromptTokens
		sumCompletion += usage.CompletionTokens
		sumTotal += usage.TotalTokens

		content = cleanupLLMJSON(content)
		var out llmOutput
		if err := json.Unmarshal([]byte(content), &out); err != nil {
			return nil, fmt.Errorf("decode llm json content (chunk %d/%d): %w; content=%s", ci+1, len(chunks), err, truncateForErr(content, 400))
		}

		groups, st := mapLLMOutputToGroups(text, chunk, span[0], model, out)
		totalLLMGroups += st.llmGroupsWithTwoFrags
		totalDroppedFrags += st.droppedFragments
		allGroups = append(allGroups, groups...)
	}

	allGroups = dedupeCloneGroups(allGroups)

	if len(allGroups) == 0 && totalLLMGroups > 0 {
		fmt.Fprintf(os.Stderr, "openai: the model returned %d groups with 2+ fragments, but none passed the substring check of the source text (%d fragments were discarded). Quotes must match the text byte-by-byte.\n",
			totalLLMGroups, totalDroppedFrags)
	} else if len(allGroups) == 0 && totalLLMGroups == 0 && len(chunks) == 1 {
		fmt.Fprintf(os.Stderr, "openai: the model returned an empty list of groups (no repeats were found or the response was cut off).\n")
	}

	cfg.CustomParams["openai_usage"] = map[string]interface{}{
		"prompt_tokens":     sumPrompt,
		"completion_tokens": sumCompletion,
		"total_tokens":      sumTotal,
		"chunks":            len(chunks),
	}

	if cfg.MinGroupPower > 0 {
		filtered := allGroups[:0]
		for _, g := range allGroups {
			if len(g.Fragments) >= cfg.MinGroupPower {
				filtered = append(filtered, g)
			}
		}
		allGroups = filtered
	}

	return allGroups, nil
}

type openaiUsageSnapshot struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func openaiChatOnce(client *http.Client, url, apiKey string, payload openAIRequest) (content string, usage openaiUsageSnapshot, err error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", usage, fmt.Errorf("marshal openai request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", usage, fmt.Errorf("create openai request: %w", err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", usage, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	rawResp, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", usage, fmt.Errorf("openai status %d: %s", resp.StatusCode, string(rawResp))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(rawResp, &parsed); err != nil {
		return "", usage, fmt.Errorf("decode openai response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", usage, fmt.Errorf("openai returned no choices")
	}

	msg, err := decodeMessageContent(parsed.Choices[0].Message.Content)
	if err != nil {
		return "", usage, fmt.Errorf("decode assistant message content: %w", err)
	}

	usage = openaiUsageSnapshot{
		PromptTokens:     parsed.Usage.PromptTokens,
		CompletionTokens: parsed.Usage.CompletionTokens,
		TotalTokens:      parsed.Usage.TotalTokens,
	}
	return msg, usage, nil
}

func decodeMessageContent(raw json.RawMessage) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return "", nil
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", err
		}
		return s, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" && p.Text != "" {
			b.WriteString(p.Text)
		}
	}
	return b.String(), nil
}

type mapLLMStats struct {
	llmGroupsWithTwoFrags int
	droppedFragments      int
}

func mapLLMOutputToGroups(fullText, chunk string, chunkByteOffset int, model string, out llmOutput) ([]framework.CloneGroup, mapLLMStats) {
	var st mapLLMStats
	groups := make([]framework.CloneGroup, 0, len(out.Groups))

	for _, g := range out.Groups {
		if len(g.Fragments) < 2 {
			continue
		}
		st.llmGroupsWithTwoFrags++

		fragments := make([]framework.TextFragment, 0, len(g.Fragments))
		searchFrom := 0
		for _, f := range g.Fragments {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}

			idx := strings.Index(chunk[searchFrom:], f)
			if idx < 0 {
				idx = strings.Index(chunk, f)
			} else {
				idx += searchFrom
			}
			if idx < 0 {
				st.droppedFragments++
				continue
			}

			globalByte := chunkByteOffset + idx
			if globalByte < 0 || globalByte > len(fullText) {
				st.droppedFragments++
				continue
			}

			start := len(strings.Fields(fullText[:globalByte]))
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

	return groups, st
}

func dedupeCloneGroups(in []framework.CloneGroup) []framework.CloneGroup {
	seen := make(map[string]struct{}, len(in))
	out := make([]framework.CloneGroup, 0, len(in))
	for _, g := range in {
		k := cloneGroupKey(g)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, g)
	}
	return out
}

func cloneGroupKey(g framework.CloneGroup) string {
	parts := make([]string, len(g.Fragments))
	for i, fr := range g.Fragments {
		parts[i] = fr.Content
	}
	sort.Strings(parts)
	return strings.Join(parts, "\x00")
}

func ensureCustomParams(cfg *framework.CloneFinderConfig) {
	if cfg.CustomParams == nil {
		cfg.CustomParams = map[string]interface{}{}
	}
}

// chunkByteSpans returns half-open byte intervals [start,end) covering text in rune chunks.
func chunkByteSpans(text string, maxRunes, overlapRunes int) [][2]int {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return nil
	}
	if n <= maxRunes {
		return [][2]int{{0, len(text)}}
	}

	step := maxRunes - overlapRunes
	if step <= 0 {
		step = maxRunes / 2
		if step <= 0 {
			step = 1
		}
	}

	var spans [][2]int
	for rStart := 0; rStart < n; {
		rEnd := rStart + maxRunes
		if rEnd > n {
			rEnd = n
		}
		b0 := runeIndexToByteOffset(text, rStart)
		b1 := runeIndexToByteOffset(text, rEnd)
		spans = append(spans, [2]int{b0, b1})
		if rEnd == n {
			break
		}
		rStart += step
		if rStart >= n {
			break
		}
	}
	return spans
}

func runeIndexToByteOffset(s string, runeIdx int) int {
	if runeIdx <= 0 {
		return 0
	}
	i, pos := 0, 0
	for pos < runeIdx && i < len(s) {
		_, sz := utf8.DecodeRuneInString(s[i:])
		if sz == 0 {
			break
		}
		i += sz
		pos++
	}
	if i > len(s) {
		return len(s)
	}
	return i
}

func truncateForErr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
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
