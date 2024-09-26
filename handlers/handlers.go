package handlers

import (
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"semantic-cache/database"
	"semantic-cache/embeddings"
)

type RequestBody struct {
	Message string `json:"user_message"`
}

type ResponseBody struct {
	MessageFound  bool                     `json:"message_found"`
	CachedPayload []database.GetOutputJSON `json:"cached_payload"`
}

type PutRequestBody struct {
	Message       string `json:"user_message"`
	ModelResponse string `json:"model_response"`
}

type PutResponseBody struct {
	Result string `json:"operation_result"`
}

func HandleGetRequest(c *fiber.Ctx) error {

	c.Accepts("text/plain", "application/json")
	c.Accepts("json", "text")

	log.Info().Msg("Handling GET request")
	// Parse the JSON body using Sonic
	var reqBody RequestBody
	if err := sonic.Unmarshal(c.Body(), &reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	log.Info().Msgf("Received request body: %v", reqBody)

	// Execute a couple of steps (example operations)
	reqBody.Message = strings.ToLower(reqBody.Message)

	log.Info().Msgf("Converted message to lowercase: %v", reqBody.Message)

	// create vectors for query
	vectors, err := embeddings.CreateEmbeddings(reqBody.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create embeddings",
		})
	}

	log.Info().Msg("Created vectors for query")

	// query qdrant for response
	// initialize databases
	qdrantClient := database.InitializeQdrant()

	log.Info().Msg("Initialized Qdrant client")

	searchResults, err := database.GetQdrant(qdrantClient, vectors)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get search results",
		})
	}

	log.Info().Msgf("Received search results: %v", searchResults)

	if len(searchResults) == 0 {

		// Prepare the response
		respBody := ResponseBody{
			CachedPayload: searchResults,
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
		CachedPayload: searchResults,
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
	
	c.Accepts("text/plain", "application/json")
	c.Accepts("json", "text")

	
	// Parse the JSON body using Sonic
	var reqBody PutRequestBody
	if err := sonic.Unmarshal(c.Body(), &reqBody); err != nil {
		log.Error().Msg("Cannot parse JSON")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	log.Info().Msgf("Received request body: %v", reqBody)

	// Execute a couple of steps (example operations)
	reqBody.Message = strings.ToLower(reqBody.Message)

	log.Info().Msgf("Converted message to lowercase: %v", reqBody.Message)

	// create vectors for query
	vectors, err := embeddings.CreateEmbeddings(reqBody.Message)
	if err != nil {
		log.Error().Msgf("Error creating vectors for query: %v", err)
	}

	log.Info().Msg("Created vectors for query")

	// query qdrant for response
	// initialize databases
	qdrantClient := database.InitializeQdrant()

	log.Info().Msg("Initialized Qdrant client")

	operationInfo := database.PutQdrant(qdrantClient, vectors, reqBody.Message, reqBody.ModelResponse)

	log.Info().Msgf("Received operation info: %v", operationInfo)

	// Prepare the response
	respBody := PutResponseBody{
		Result: "operationInfo",
	}

	// Encode the response using Sonic
	jsonResp, err := sonic.Marshal(respBody)
	if err != nil {
		log.Error().Msg("Failed to encode response")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encode response",
		})
	}

	// Set content type and send the response
	c.Set("Content-Type", "application/json")
	return c.Send(jsonResp)
}
