package database

import (
	"context" // understand this and usage in file
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"

	"github.com/rs/zerolog/log"
)

var collectionName = os.Getenv("QDRANT_COLLECTION")

type ScoredPoint struct {
	Id      *qdrant.PointId          `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"` // Point id
	Payload map[string]*qdrant.Value `json:"payload,omitempty"`                                  // Payload
	/* 155-byte string literal not displayed */
	Score      float32            `protobuf:"fixed32,3,opt,name=score,proto3" json:"score,omitempty"`                                 // Similarity score
	Version    uint64             `protobuf:"varint,5,opt,name=version,proto3" json:"version,omitempty"`                              // Last update operation applied to this point
	Vectors    *qdrant.Vectors    `protobuf:"bytes,6,opt,name=vectors,proto3,oneof" json:"vectors,omitempty"`                         // Vectors to search
	ShardKey   *qdrant.ShardKey   `protobuf:"bytes,7,opt,name=shard_key,json=shardKey,proto3,oneof" json:"shard_key,omitempty"`       // Shard key
	OrderValue *qdrant.OrderValue `protobuf:"bytes,8,opt,name=order_value,json=orderValue,proto3,oneof" json:"order_value,omitempty"` // Order by value
	// contains filtered or unexported fields
}

type GetOutputJSON struct {
	Score         float32 `json:"score"`
	UserMessage   string  `json:"user_message"`
	ModelResponse string  `json:"model_response"`
}

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
		log.Fatal().Msgf("Could not get health: %v", err)
	}
	log.Printf("Qdrant version: %s", healthCheckResult.GetVersion())

	// check if collection exists
	exists, err := client.CollectionExists(context.Background(), collectionName)
	if err != nil {
		log.Fatal().Msgf("Could not check if collection exists: %v", err)
	}
	if exists {
		log.Info().Msgf("Collection %s exists", collectionName)
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
		log.Fatal().Msgf("Could not create collection: %v", err)
	} else {
		log.Info().Msgf("Collection %s created", collectionName)
	}

	return client
}

func GetQdrant(client *qdrant.Client, vectors []float32) ([]GetOutputJSON, error) {
	// Query the database
	searchedPoints, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryDense(vectors),
		WithPayload:    qdrant.NewWithPayloadInclude("model_response", "user_message"),
		ScoreThreshold: qdrant.PtrOf(float32(0.7)), // TODO: make this configurable
	})
	if err != nil {
		log.Fatal().Msgf("Could not search points: %v", err)
	}

	client.Close()

	log.Info().Msg("Searched points")

	var outputData []GetOutputJSON
	for _, item := range searchedPoints {
		output := GetOutputJSON{
			Score:         item.Score,
			UserMessage:   item.Payload["user_message"].GetStringValue(),
			ModelResponse: item.Payload["model_response"].GetStringValue(),
		}
		outputData = append(outputData, output)
	}

	return outputData, err
}

func PutQdrant(client *qdrant.Client, vectors []float32, message string, modelResponse string) *qdrant.UpdateResult {
	id, _ := uuid.NewRandom()

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

	log.Info().Msgf("Upserting %d points", len(upsertPoints))

	operationInfo, err := client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         upsertPoints,
	})
	if err != nil {
		log.Fatal().Msgf("Could not upsert points: %v", err)
	}
	fmt.Println("Upsert", len(upsertPoints), "points")

	client.Close()

	return operationInfo
}
