package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func debugMongo() {
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

	// List all databases using a simpler approach
	result, err := client.ListDatabases(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to list databases: %v", err)
	}

	var databases []string
	for _, db := range result.Databases {
		databases = append(databases, db.Name)
	}

	log.Printf("Available databases: %v", databases)

	// Check each database for movie collections
	for _, dbName := range databases {
		log.Printf("\n=== Checking database: %s ===", dbName)
		db := client.Database(dbName)

		collections, err := db.ListCollectionNames(context.Background(), nil)
		if err != nil {
			log.Printf("Failed to list collections in %s: %v", dbName, err)
			continue
		}
		log.Printf("Collections in %s: %v", dbName, collections)

		// Check for movies collection and count documents
		for _, collectionName := range collections {
			if collectionName == "movies" {
				collection := db.Collection(collectionName)
				count, err := collection.CountDocuments(context.Background(), nil)
				if err != nil {
					log.Printf("Failed to count documents in %s.%s: %v", dbName, collectionName, err)
				} else {
					log.Printf("Found %d documents in %s.%s", count, dbName, collectionName)

					if count > 0 {
						// Get a sample document
						var sample interface{}
						err := collection.FindOne(context.Background(), nil).Decode(&sample)
						if err == nil {
							log.Printf("Sample document: %+v", sample)
						}
					}
				}
			}
		}
	}
}

func main() {
	debugMongo()
}
