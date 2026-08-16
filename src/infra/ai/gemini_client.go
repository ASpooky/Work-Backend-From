package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// GeminiClient talks to the Gemini generateContent REST API directly (no
// CLI subprocess): it accepts an API key via GEMINI_API_KEY at
// construction time, not a browser login, so it works headless in a
// backend process.
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		baseURL:    defaultBaseURL,
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type generationConfig struct {
	ResponseMimeType string         `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]any `json:"responseSchema,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent   `json:"contents"`
	SystemInstruction *geminiContent    `json:"systemInstruction,omitempty"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

// Chat sends the conversation history and returns the model's free-text
// reply, for the back-and-forth "壁打ち" goal-clarifying conversation.
func (c *GeminiClient) Chat(ctx context.Context, systemInstruction string, messages []entity.ChatMessage) (string, error) {
	return c.call(ctx, systemInstruction, messages, nil)
}

// GenerateJSON sends the conversation history plus a system instruction and
// a Gemini-format JSON schema (OpenAPI subset: type/properties/items/...),
// returning the raw JSON text the model produced. The caller is
// responsible for unmarshaling it into a domain-specific shape.
func (c *GeminiClient) GenerateJSON(ctx context.Context, systemInstruction string, messages []entity.ChatMessage, schema map[string]any) (string, error) {
	return c.call(ctx, systemInstruction, messages, schema)
}

func (c *GeminiClient) call(ctx context.Context, systemInstruction string, messages []entity.ChatMessage, schema map[string]any) (string, error) {
	contents := make([]geminiContent, len(messages))
	for i, m := range messages {
		contents[i] = geminiContent{Role: string(m.Role), Parts: []geminiPart{{Text: m.Content}}}
	}

	reqBody := geminiRequest{Contents: contents}
	if systemInstruction != "" {
		reqBody.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: systemInstruction}}}
	}
	if schema != nil {
		reqBody.GenerationConfig = &generationConfig{ResponseMimeType: "application/json", ResponseSchema: schema}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini API returned no candidates")
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
