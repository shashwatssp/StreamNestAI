package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	log.Printf("Connecting to MongoDB with URI: %s", mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Test connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Successfully connected to MongoDB")

	// Check specific databases that might contain movies
	databasesToCheck := []string{"streamnestai", "StreamNestAI", "admin", "test", "movies"}

	for _, dbName := range databasesToCheck {
		log.Printf("\n=== Checking database: %s ===", dbName)
		db := client.Database(dbName)

		// Try to access the movies collection directly
		collection := db.Collection("movies")
		count, err := collection.CountDocuments(context.Background(), nil)
		if err != nil {
			log.Printf("Failed to count documents in %s.movies: %v", dbName, err)
		} else {
			log.Printf("Found %d documents in %s.movies", count, dbName, dbName)

			if count > 0 {
				// Get a sample document
				var sample map[string]interface{}
				err := collection.FindOne(context.Background(), nil).Decode(&sample)
				if err == nil {
					log.Printf("Sample document structure: %+v", sample)
				}
			}
		}
	}

	// Try to list collections in streamnestai database
	log.Printf("\n=== Listing collections in streamnestai database ===")
	db := client.Database("streamnestai")

	// Use RunCommand to list collections
	var result struct {
		Collections []struct {
			Name string `bson:"name"`
			Type string `bson:"type"`
		} `bson:"collections"`
	}

	err = db.RunCommand(context.Background(), map[string]interface{}{
		"listCollections": 1,
	}).Decode(&result)

	if err != nil {
		log.Printf("Failed to list collections: %v", err)
	} else {
		log.Printf("Collections in streamnestai: %v", result.Collections)
	}
}
