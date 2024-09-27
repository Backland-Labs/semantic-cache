package main

import (
	"os"

	"semantic-cache/handlers"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Message struct {
	Text string `json:"text"`
}

func main() {
	// configure logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	// Provide a minimal config
	app.Use(healthcheck.New())
	app.Use(idempotency.New())

	// Check the cache to see if there is any data

	app.Get("/get", handlers.HandleGetRequest)

	// Put data in the cache
	app.Post("/post", handlers.HandlePutRequest)

	log.Info().Msg("Server starting on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
