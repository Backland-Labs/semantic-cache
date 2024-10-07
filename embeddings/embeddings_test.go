package embeddings

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
)

// Modify CreateOpenAIEmbeddings to accept a custom URL

func TestCreateOpenAIEmbeddings(t *testing.T) {
	// Set up test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test_api_key" {
			t.Errorf("Expected Authorization: Bearer test_api_key, got %s", r.Header.Get("Authorization"))
		}

		// Check request body
		body, _ := io.ReadAll(r.Body)
		var reqBody EmbeddingRequest
		err := sonic.Unmarshal(body, &reqBody)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}
		if reqBody.Input != "test input" {
			t.Errorf("Expected input 'test input', got '%s'", reqBody.Input)
		}
		if reqBody.Model != "text-embedding-3-small" {
			t.Errorf("Expected model 'text-embedding-3-small', got '%s'", reqBody.Model)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := EmbeddingResponse{
			Object: "list",
			Data: []struct {
				Object    string    `json:"object"`
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{
					Object:    "embedding",
					Embedding: []float32{0.1, 0.2, 0.3},
					Index:     0,
				},
			},
			Model: "text-embedding-3-small",
			Usage: struct {
				PromptTokens int `json:"prompt_tokens"`
				TotalTokens  int `json:"total_tokens"`
			}{
				PromptTokens: 5,
				TotalTokens:  5,
			},
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Set environment variables
	os.Setenv("OPENAI_API_KEY", "test_api_key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Call the function with the test server URL
	embeddings, err := CreateOpenAIEmbeddings("test input", server.URL)
	// Check results
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := []float32{0.1, 0.2, 0.3}
	if !reflect.DeepEqual(embeddings, expected) {
		t.Errorf("Expected embeddings %v, got %v", expected, embeddings)
	}
}

func TestCreateLocalEmbeddings(t *testing.T) {
	// Set up test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Check request body
		body, _ := io.ReadAll(r.Body)
		var reqBody LocalEmbeddingRequest
		err := sonic.Unmarshal(body, &reqBody)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}
		if reqBody.Prompt != "test input" {
			t.Errorf("Expected prompt 'test input', got '%s'", reqBody.Prompt)
		}
		if reqBody.Model != "nomic-embed-text" {
			t.Errorf("Expected model 'nomic-embed-text', got '%s'", reqBody.Model)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := LocalEmbeddingReponse{
			Embeddings: []float32{0.1, 0.2, 0.3},
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Call the function with the test server URL
	embeddings, err := CreateLocalEmbeddings("test input", server.URL)
	// Check results
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := []float32{0.1, 0.2, 0.3}
	if !reflect.DeepEqual(embeddings, expected) {
		t.Errorf("Expected embeddings %v, got %v", expected, embeddings)
	}
}

func TestCreateOpenAIEmbeddings_ErrorCases(t *testing.T) {
	// Test missing API key
	os.Unsetenv("OPENAI_API_KEY")
	_, err := CreateOpenAIEmbeddings("test input", "")
	if err == nil || err.Error() != "OPENAI_API_KEY environment variable not set" {
		t.Errorf("Expected error for missing API key, got %v", err)
	}

	// Test server error
	os.Setenv("OPENAI_API_KEY", "test_api_key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err = CreateOpenAIEmbeddings("test input", server.URL)
	if err == nil {
		t.Errorf("Expected error for server error, got nil")
	}
}

func TestCreateLocalEmbeddings_ErrorCases(t *testing.T) {
	// Test server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := CreateLocalEmbeddings("test input", server.URL)
	if err == nil {
		t.Errorf("Expected error for server error, got nil")
	}

	// Test empty response
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = json.NewEncoder(w).Encode(LocalEmbeddingReponse{Embeddings: []float32{}})
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	_, err = CreateLocalEmbeddings("test input", server.URL)
	if err == nil || err.Error() != "no embedding data in response" {
		t.Errorf("Expected error for empty response, got %v", err)
	}
}
