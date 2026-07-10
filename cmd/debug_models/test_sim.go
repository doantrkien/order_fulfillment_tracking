package main

import (
	"context"
	"fmt"
	"main/pkg/embedding"
	"os"

	"github.com/pgvector/pgvector-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	os.Setenv("OLLAMA_BASE_URL", "http://localhost:11434")
	os.Setenv("OLLAMA_EMBEDDING_MODEL", "nomic-embed-text-v2-moe")

	client, err := embedding.NewOllamaEmbeddingClient()
	if err != nil {
		panic(err)
	}

	query := "search_query: send to the security"
	vec, err := client.Embed(context.Background(), query)
	if err != nil {
		panic(err)
	}

	dsn := "host=localhost user=postgres password=12345 dbname=order_tracking port=5435 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	type Result struct {
		Heading string
		Sim     float64
	}
	var results []Result

	pgVec := pgvector.NewVector(vec)
	err = db.Raw(`
		SELECT heading, 1 - (embedding <=> ?) as sim
		FROM knowledge_chunks
		ORDER BY embedding <=> ?
		LIMIT 10
	`, pgVec, pgVec).Scan(&results).Error
	if err != nil {
		panic(err)
	}

	fmt.Println("Query:", query)
	fmt.Println("Top 10 chunks:")
	for _, r := range results {
		fmt.Printf("Sim: %.4f | Heading: %s\n", r.Sim, r.Heading)
	}
}
