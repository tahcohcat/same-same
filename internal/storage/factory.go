package storage

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/tahcohcat/same-same/internal/storage/local"
	"github.com/tahcohcat/same-same/internal/storage/memory"
)

// NewStorageFromEnv returns a Storage implementation based on STORAGE_TYPE env var
func NewStorageFromEnv() (Storage, error) {
	_ = godotenv.Load() // load .env if present
	typeStr := os.Getenv("STORAGE_TYPE")
	if typeStr == "local" {
		basePath := os.Getenv("LOCAL_STORAGE_PATH")
		if basePath == "" {
			basePath = "./data/storage" // default path
		}
		collection := os.Getenv("STORAGE_COLLECTION")
		if collection == "" {
			collection = "default" // default collection name
		}

		return local.NewVectorStorageAdapter(basePath, collection)
	}
	// default to memory
	return memory.NewStorage(), nil
}
