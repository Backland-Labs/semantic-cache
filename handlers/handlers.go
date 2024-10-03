package handlers

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"semantic-cache/database"
	"semantic-cache/embeddings"
)

var validate = validator.New()

type RequestBody struct {
	Message string `json:"user_message" validate:"required,min=1"`
}

type ResponseBody struct {
	MessageFound  bool                     `json:"message_found"`
	CachedPayload []database.GetOutputJSON `json:"cached_payload"`
}

type PutRequestBody struct {
	Message       string `json:"user_message" validate:"required,min=1"`
	ModelResponse string `json:"model_response" validate:"required,min=1"`
}

type PutResponseBody struct {
	Result string `json:"result"`
}

func HandleGetRequest(c *fiber.Ctx) error {
	c.Accepts("text/plain", "application/json")

	log.Info().Msg("Handling GET request")
	// Parse the JSON body using Sonic
	var reqBody RequestBody
	if err := c.App().Config().JSONDecoder(c.Body(), &reqBody); err != nil {
		log.Error().Msg("cannot parse JSON")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	if err := validate.Struct(reqBody); err != nil {
		log.Error().Msg("invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	log.Info().Msgf("Received request body: %v", reqBody)

	// Execute a couple of steps (example operations)
	reqBody.Message = strings.ToLower(reqBody.Message)

	log.Info().Msgf("Converted message to lowercase: %v", reqBody.Message)

	// create vectors for query
	vectors, err := embeddings.CreateLocalEmbeddings(reqBody.Message)
	if err != nil {
		log.Error().Msgf("error creating vectors for query: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create embeddings",
		})
	}

	log.Info().Msg("Created vectors for query")

	// query qdrant for response
	// initialize databases
	qdrantClient := database.GetQdrantClient() // specificy this only once

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
		jsonResp, err := c.App().Config().JSONEncoder(respBody)
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
	jsonResp, err := c.App().Config().JSONEncoder(respBody)
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

	// Parse the JSON body using Sonic
	var reqBody PutRequestBody
	if err := c.App().Config().JSONDecoder(c.Body(), &reqBody); err != nil {
		log.Error().Msg("Cannot parse JSON")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if err := validate.Struct(reqBody); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Info().Msgf("Received request body: %v", reqBody)

	// convert to lowercase
	reqBody.Message = strings.ToLower(reqBody.Message)

	log.Info().Msgf("Converted message to lowercase: %v", reqBody.Message)

	// create vectors for query
	vectors, err := embeddings.CreateLocalEmbeddings(reqBody.Message)
	if err != nil {
		log.Error().Msgf("Error creating vectors for query: %v", err)
	}

	log.Info().Msg("Created vectors for query")

	// Acess Qdrant client
	qdrantClient := database.GetQdrantClient()

	operationInfo := database.PutQdrant(qdrantClient, vectors, reqBody.Message, reqBody.ModelResponse)

	log.Info().Msgf("received operation info: %v", operationInfo)

	// Prepare the response
	respBody := PutResponseBody{
		Result: operationInfo.String(),
	}

	// Encode the response
	jsonResp, err := c.App().Config().JSONEncoder(respBody)
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
