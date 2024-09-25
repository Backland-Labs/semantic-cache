package database

import (
	"context" // understand this and usage in file
	"fmt"
	"log" // swap for other logging library
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

var (
	collectionName              = "test_collection"
	vectorSize           uint64 = 4
	distance                    = qdrant.Distance_Dot
	defaultSegmentNumber uint64 = 2
)

func InitializeQdrant() *qdrant.Client {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   os.Getenv("QDRANT_HOST"),
		Port:   6334,
		UseTLS: false,
	})
	if err != nil {
		panic(err)
	}

	// Get a context for a minute
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Execute health check
	healthCheckResult, err := client.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("Could not get health: %v", err)
	}
	log.Printf("Qdrant version: %s", healthCheckResult.GetVersion())

	// check if collection exists
	exists, err := client.CollectionExists(context.Background(), collectionName)
	if err != nil {
		log.Fatalf("Could not check if collection exists: %v", err)
	}
	if exists {
		log.Println("Collection", collectionName, "exists")
		return client
	}

	err = client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     1536,
			Distance: qdrant.Distance_Cosine,
			OnDisk:   qdrant.PtrOf(true),
		}),
		QuantizationConfig: qdrant.NewQuantizationScalar(&qdrant.ScalarQuantization{
			Type:      qdrant.QuantizationType_Int8,
			AlwaysRam: qdrant.PtrOf(true),
		}),
	})
	if err != nil {
		log.Fatalf("Could not create collection: %v", err)
	} else {
		log.Println("Collection", collectionName, "created")
	}

	return client
}

func GetQdrant(client *qdrant.Client, vectors []float32) ([]*qdrant.ScoredPoint, error) {
	// Query the database
	searchedPoints, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryDense(vectors),
		WithPayload:    qdrant.NewWithPayloadInclude("model_response"),
	})
	if err != nil {
		log.Fatalf("Could not search points: %v", err)
	}

	client.Close()

	log.Printf("Found points: %s", searchedPoints)

	return searchedPoints, err
}

func PutQdrant(client *qdrant.Client, vectors []float32, message string, modelResponse string) *qdrant.UpdateResult {
	id, err := uuid.NewRandom()

	// Upsert some data
	waitUpsert := true
	upsertPoints := []*qdrant.PointStruct{
		{
			Id:      qdrant.NewIDUUID(id.String()),
			Vectors: qdrant.NewVectorsDense(vectors),
			Payload: qdrant.NewValueMap(map[string]any{
				"user_message":   message,
				"model_response": modelResponse,
			}),
		},
	}

	log.Println("Upsert", len(upsertPoints), "points")

	operationInfo, err := client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         upsertPoints,
	})
	if err != nil {
		log.Fatalf("Could not upsert points: %v", err)
	}
	fmt.Println("Upsert", len(upsertPoints), "points")

	client.Close()

	return operationInfo
}
