package main

import (
	"os"

	"semantic-cache/handlers"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
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

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Get("/get", handlers.HandleGetRequest)
	app.Post("/post", handlers.HandlePutRequest)
	// app.Get("/health")

	log.Info().Msg("Server starting on :8080")
	app.Listen(":8080")
	// log.Fatal(app.Listen(":8080"))
}
