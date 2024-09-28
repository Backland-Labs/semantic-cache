package database

import (
	"context" // understand this and usage in file
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"

	"github.com/rs/zerolog/log"
)

var (
	qdrantClientInstance *qdrant.Client
	qdrantClientOnce     sync.Once
	collectionName       = os.Getenv("QDRANT_COLLECTION")
	qdrantHost           = os.Getenv("QDRANT_HOST")
)

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

func initializeQdrant() (*qdrant.Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   qdrantHost,
		Port:   6334,
		UseTLS: false,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	healthCheckResult, err := client.HealthCheck(ctx)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("Qdrant version: %s", healthCheckResult.GetVersion())

	exists, err := client.CollectionExists(context.Background(), collectionName)
	if err != nil {
		return nil, err
	}

	if exists {
		log.Info().Msgf("Collection %s exists", collectionName)
		return client, nil
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
		OptimizersConfig: &qdrant.OptimizersConfigDiff{
			DefaultSegmentNumber: qdrant.PtrOf(uint64(16)), // used to minimize latency set to 2 to maximize throughput
		},
	})
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Collection %s created", collectionName)
	return client, nil
}

// GetQdrantClient returns a singleton instance of the Qdrant client
func GetQdrantClient() *qdrant.Client {
	qdrantClientOnce.Do(func() {
		var err error
		qdrantClientInstance, err = initializeQdrant()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Qdrant client")
		}
	})
	return qdrantClientInstance
}

func GetQdrant(client *qdrant.Client, vectors []float32) ([]GetOutputJSON, error) {
	// Query the database
	searchedPoints, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryDense(vectors),
		WithPayload:    qdrant.NewWithPayloadInclude("model_response", "user_message"),
		ScoreThreshold: qdrant.PtrOf(float32(0.7)), // TODO: make this configurable
		Params: &qdrant.SearchParams{
			Quantization: &qdrant.QuantizationSearchParams{
				Rescore: qdrant.PtrOf(true), // remove if results are inaccruate
			},
		},
	})
	if err != nil {
		log.Fatal().Msgf("Could not search points: %v", err)
	}

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

	return operationInfo
}
