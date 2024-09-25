package handlers

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/qdrant/go-client/qdrant"
	"semantic-cache/database"
	"semantic-cache/embeddings"
)

type RequestBody struct {
	Message string `json:"user_message"`
}

type ResponseBody struct {
	CachedMessage []*qdrant.ScoredPoint `json:"cached_message"`
	MessageFound  bool                  `json:"message_found"`
}

type PutRequestBody struct {
	Message       string `json:"user_message"`
	ModelResponse string `json:"model_response"`
}

type PutResponseBody struct {
	Result string `json:"operation_result"`
}

func HandleGetRequest(c *fiber.Ctx) error {
	// Parse the JSON body using Sonic
	var reqBody RequestBody
	if err := sonic.Unmarshal(c.Body(), &reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	fmt.Println(reqBody)

	// Execute a couple of steps (example operations)
	reqBody.Message = strings.ToLower(reqBody.Message)

	// create vectors for query
	vectors, err := embeddings.CreateEmbeddings(reqBody.Message)

	// query qdrant for response
	// initialize databases
	qdrantClient := database.InitializeQdrant()

	searchResults, err := database.GetQdrant(qdrantClient, vectors)

	fmt.Println(searchResults)

	if len(searchResults) == 0 {

		// Prepare the response
		respBody := ResponseBody{
			CachedMessage: searchResults,
			MessageFound:  false,
		}

		// Encode the response using Sonic
		jsonResp, err := sonic.Marshal(respBody)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to encode response",
			})
		}

		// Set content type and send the response
		c.Set("Content-Type", "application/json")
		return c.Send(jsonResp)
	}

	// Prepare the response
	respBody := ResponseBody{
		CachedMessage: searchResults,
		MessageFound:  true,
	}

	// Encode the response using Sonic
	jsonResp, err := sonic.Marshal(respBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encode response",
		})
	}

	// Set content type and send the response
	c.Set("Content-Type", "application/json")
	return c.Send(jsonResp)
}

func HandlePutRequest(c *fiber.Ctx) error {
	// Parse the JSON body using Sonic
	var reqBody PutRequestBody
	if err := sonic.Unmarshal(c.Body(), &reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	fmt.Println(reqBody)

	// Execute a couple of steps (example operations)
	reqBody.Message = strings.ToLower(reqBody.Message)

	// create vectors for query
	vectors, err := embeddings.CreateEmbeddings(reqBody.Message)

	// query qdrant for response
	// initialize databases
	qdrantClient := database.InitializeQdrant()

	operationInfo := database.PutQdrant(qdrantClient, vectors, reqBody.Message, reqBody.ModelResponse)

	fmt.Println(operationInfo)

	// Prepare the response
	respBody := PutResponseBody{
		Result: "operationInfo",
	}

	// Encode the response using Sonic
	jsonResp, err := sonic.Marshal(respBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encode response",
		})
	}

	// Set content type and send the response
	c.Set("Content-Type", "application/json")
	return c.Send(jsonResp)
}
