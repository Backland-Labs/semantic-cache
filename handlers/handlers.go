package handlers

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"semantic-cache/embeddings"
	"semantic-cache/database"
)

type RequestBody struct {
	Message string `json:"user_message"`
}

type ResponseBody struct {
	ProcessedMessage string `json:"processed_message"`
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

	// Prepare the response
	respBody := ResponseBody{
		ProcessedMessage: "hi",
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
