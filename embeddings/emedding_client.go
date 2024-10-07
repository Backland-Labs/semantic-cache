package embeddings

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

type LocalEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type LocalEmbeddingReponse struct {
	Embeddings []float32 `json:"embedding"`
}

type EmbeddingRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format"`
}

type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func CreateOpenAIEmbeddings(input string, url string) ([]float32, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal().Msg("OPENAI_API_KEY environment variable not set")
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	requestBody := EmbeddingRequest{
		Input:          input,
		Model:          "text-embedding-3-small",
		EncodingFormat: "float",
	}

	jsonData, err := sonic.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}

	log.Info().Msgf("Creating request body: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	log.Info().Msgf("Sending request to %s", url)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	log.Info().Msgf("Response status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var embeddingResponse EmbeddingResponse
	err = sonic.Unmarshal(body, &embeddingResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if len(embeddingResponse.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embeddingResponse.Data[0].Embedding, nil
}

func CreateLocalEmbeddings(input string, url string) ([]float32, error) {
	requestBody := LocalEmbeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: input,
	}

	jsonData, err := sonic.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}

	log.Info().Msgf("Creating request body: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Info().Msgf("Sending request to %s", url)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	log.Info().Msgf("Response status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var embeddingResponse LocalEmbeddingReponse
	err = sonic.Unmarshal(body, &embeddingResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if len(embeddingResponse.Embeddings) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embeddingResponse.Embeddings, nil
}
