package main

import (
	"log"

	"semantic-cache/handlers"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

type Message struct {
	Text string `json:"text"`
}

func main() {

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Get("/get", handlers.HandleGetRequest)
	//app.Post("/post", database.PostQdrant(qdrantClient))
	//app.Get("/health")

	log.Println("Server starting on :8080")
	log.Fatal(app.Listen(":8080"))
}
