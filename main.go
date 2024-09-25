package main

import (
	"os"

	"semantic-cache/handlers"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
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

	app.Get("/get", handlers.HandleGetRequest)
	app.Post("/post", handlers.HandlePutRequest)

	log.Info().Msg("Server starting on :8080")
	app.Listen(":8080")
	// log.Fatal(app.Listen(":8080"))
}
