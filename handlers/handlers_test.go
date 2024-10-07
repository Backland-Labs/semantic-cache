package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	MockExternalServer struct {
		mock.Mock
	}

	MockEmbeddings struct {
		mock.Mock
	}

	MockQdrantClient struct {
		mock.Mock
	}

	MockValidator struct {
		mock.Mock
	}

)


// Mock method implementations
func (m *MockExternalServer) Call(message string) (bool, error) {
	args := m.Called(message)
	return args.Bool(0), args.Error(1)
}

func (m *MockEmbeddings) CreateLocalEmbeddings(input string) ([]float32, error) {
	args := m.Called(input)
	return args.Get(0).([]float32), args.Error(1)
}

func (m *MockQdrantClient) Query(ctx context.Context, req *qdrant.QueryPoints) ([]*qdrant.ScoredPoint, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*qdrant.ScoredPoint), args.Error(1)
}

func (m *MockValidator) Struct(s interface{}) error {
	args := m.Called(s)
	return args.Error(0)
}

func setupTest() (*fiber.App, *MockExternalServer, *MockEmbeddings, *MockQdrantClient, *MockValidator) {
	app := fiber.New()
	mockServer := &MockExternalServer{}
	mockEmbeddings := new(MockEmbeddings)
	mockQdrantClient := new(MockQdrantClient)
	mockValidator := new(MockValidator)

	app.Get("/", HandleGetRequest)
	app.Put("/", HandlePutRequest)

	return app, mockServer, mockEmbeddings, mockQdrantClient, mockValidator
}

func createRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		assert.NoError(t, err, "Failed to marshal JSON")
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	return req
}

func TestHandleGetRequest(t *testing.T) {
	app, _, _, _, _ := setupTest()

	t.Run("Missing required fields", func(t *testing.T) {
		req := createRequest(t, "GET", "/", map[string]string{"message": "some-message"})
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Error creating local embeddings", func(t *testing.T) {
		req := createRequest(t, "GET", "/", map[string]string{"user_message": "hello"})
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})
}

