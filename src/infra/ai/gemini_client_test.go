package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestGeminiClient_Chat(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "test-key" {
			t.Errorf("request key = %q, want test-key", r.URL.Query().Get("key"))
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"こんにちは"}]}}]}`))
	}))
	defer server.Close()

	client := &GeminiClient{apiKey: "test-key", model: "test-model", httpClient: server.Client(), baseURL: server.URL}

	got, err := client.Chat(context.Background(), "あなたは親切なアシスタントです。", []entity.ChatMessage{
		{Role: entity.ChatRoleUser, Content: "こんにちは"},
	})
	if err != nil {
		t.Fatalf("Chat() returned unexpected error: %v", err)
	}
	if got != "こんにちは" {
		t.Errorf("Chat() = %q, want %q", got, "こんにちは")
	}

	contents, ok := capturedBody["contents"].([]any)
	if !ok || len(contents) != 1 {
		t.Fatalf("request contents = %v, want 1 entry", capturedBody["contents"])
	}
	if _, hasSystem := capturedBody["systemInstruction"]; !hasSystem {
		t.Errorf("Chat() request missing systemInstruction")
	}
	if _, hasSchema := capturedBody["generationConfig"]; hasSchema {
		t.Errorf("Chat() request included generationConfig, want none for plain chat")
	}
}

func TestGeminiClient_GenerateJSON(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"{\"title\":\"Run 5km\"}"}]}}]}`))
	}))
	defer server.Close()

	client := &GeminiClient{apiKey: "test-key", model: "test-model", httpClient: server.Client(), baseURL: server.URL}

	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{"title": map[string]any{"type": "string"}},
	}

	got, err := client.GenerateJSON(context.Background(), "You are a planner.", []entity.ChatMessage{
		{Role: entity.ChatRoleUser, Content: "Plan my goal"},
	}, schema)
	if err != nil {
		t.Fatalf("GenerateJSON() returned unexpected error: %v", err)
	}
	if got != `{"title":"Run 5km"}` {
		t.Errorf("GenerateJSON() = %q, want %q", got, `{"title":"Run 5km"}`)
	}

	genConfig, ok := capturedBody["generationConfig"].(map[string]any)
	if !ok {
		t.Fatalf("request missing generationConfig")
	}
	if genConfig["responseMimeType"] != "application/json" {
		t.Errorf("responseMimeType = %v, want application/json", genConfig["responseMimeType"])
	}
	if _, ok := capturedBody["systemInstruction"]; !ok {
		t.Errorf("request missing systemInstruction")
	}
}

func TestGeminiClient_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"invalid API key"}}`))
	}))
	defer server.Close()

	client := &GeminiClient{apiKey: "bad-key", model: "test-model", httpClient: server.Client(), baseURL: server.URL}

	_, err := client.Chat(context.Background(), "", []entity.ChatMessage{{Role: entity.ChatRoleUser, Content: "hi"}})
	if err == nil {
		t.Fatal("Chat() with a 401 response returned nil error, want non-nil")
	}
}
