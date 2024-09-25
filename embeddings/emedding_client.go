package embeddings

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	//"os"

	"github.com/bytedance/sonic"
)

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

func CreateEmbeddings(input string) ([]float32, error) {
	apiKey := "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"
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	url := "https://api.openai.com/v1/embeddings"

	requestBody := EmbeddingRequest{
		Input:          input,
		Model:          "text-embedding-3-small",
		EncodingFormat: "float",
	}

	jsonData, err := sonic.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}

	fmt.Println(string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Print("Response status: ", resp.Status)

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
