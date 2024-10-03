package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExternalServer is a mock for the external server
type MockExternalServer struct {
	mock.Mock
}

func (m *MockExternalServer) Call(message string) (bool, error) {
	args := m.Called(message)
	return args.Bool(0), args.Error(1)
}

func TestHandleGetRequest(t *testing.T) {
	// Create a mock external server
	mockServer := &MockExternalServer{}
	mockServer.On("Call", "some-message").Return(true, nil)

	app := fiber.New()
	app.Use(logger.New())

	// Register the HandleGetRequest handler
	app.Get("/", HandleGetRequest)

	// Create a request to the root URL with no data sent
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the mock external server on the request context
	type contextKey string

	const externalServerKey contextKey = "externalServer"

	ctx := req.Context()
	ctx = context.WithValue(ctx, externalServerKey, mockServer)
	req = req.WithContext(ctx)

	// Test the request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	// Assert the response status code
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	// Test missing required fields in request body
	data := map[string]string{
		"message": "some-message",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err = http.NewRequest("GET", "/", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Test the request
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	// Assert the response status code
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	t.Run("Test HandleGetRequest with error creating local embeddings", func(t *testing.T) {
		data := map[string]string{
			"user_message": "hello",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("GET", "/", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})
}
