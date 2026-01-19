package main

import (
	"context"
	"log"

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

	// Connect to MongoDB
	mongoURI := "mongodb://localhost:27017"
	log.Printf("Connecting to MongoDB with URI: %s", mongoURI)

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

	// Check the StreamNest database specifically
	db := client.Database("StreamNest")
	log.Printf("=== Checking StreamNest database ===")

	// Try to list collections using a different approach
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
		log.Printf("Failed to list collections with listCollections: %v", err)

		// Try alternative approach
		log.Println("Trying alternative approach to list collections...")
		cursor, err := db.ListCollections(context.Background(), nil)
		if err != nil {
			log.Printf("Failed to list collections with ListCollections: %v", err)
		} else {
			var collections []map[string]interface{}
			if err := cursor.All(context.Background(), &collections); err != nil {
				log.Printf("Failed to decode collections: %v", err)
			} else {
				log.Printf("Collections found: %v", collections)
			}
		}
	} else {
		log.Printf("Collections in StreamNest: %v", result.Collections)

		// Check each collection for movies
		for _, coll := range result.Collections {
			log.Printf("Checking collection: %s", coll.Name)
			collection := db.Collection(coll.Name)
			count, err := collection.CountDocuments(context.Background(), nil)
			if err != nil {
				log.Printf("Failed to count documents in %s: %v", coll.Name, err)
			} else {
				log.Printf("Found %d documents in %s", count, coll.Name)

				if count > 0 {
					// Get a sample document
					var sample map[string]interface{}
					err := collection.FindOne(context.Background(), nil).Decode(&sample)
					if err == nil {
						log.Printf("Sample document from %s: %+v", coll.Name, sample)
					}
				}
			}
		}
	}

	// Also try to directly access common collection names
	commonCollections := []string{"movies", "Movies", "movie", "films", "content"}
	for _, collName := range commonCollections {
		log.Printf("Trying direct access to collection: %s", collName)
		collection := db.Collection(collName)
		count, err := collection.CountDocuments(context.Background(), nil)
		if err != nil {
			log.Printf("Collection %s not accessible: %v", collName, err)
		} else {
			log.Printf("*** FOUND %d documents in StreamNest.%s ***", count, collName)

			if count > 0 {
				var sample map[string]interface{}
				err := collection.FindOne(context.Background(), nil).Decode(&sample)
				if err == nil {
					log.Printf("Sample movie: %+v", sample)
				}
			}
		}
	}
}
