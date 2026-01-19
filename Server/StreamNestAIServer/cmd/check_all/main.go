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

	// Connect to MongoDB without specifying database
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

	// Use admin command to list all databases
	var result struct {
		Databases []struct {
			Name       string `bson:"name"`
			SizeOnDisk int64  `bson:"sizeOnDisk"`
			Empty      bool   `bson:"empty"`
		} `bson:"databases"`
	}

	err = client.Database("admin").RunCommand(context.Background(), map[string]interface{}{
		"listDatabases": 1,
	}).Decode(&result)

	if err != nil {
		log.Fatalf("Failed to list databases: %v", err)
	}

	log.Printf("Found %d databases:", len(result.Databases))
	for _, db := range result.Databases {
		log.Printf("- %s (size: %d bytes, empty: %v)", db.Name, db.SizeOnDisk, db.Empty)

		// Check if this database has a movies collection
		if !db.Empty {
			dbInstance := client.Database(db.Name)
			collections, err := dbInstance.ListCollectionNames(context.Background(), nil)
			if err != nil {
				log.Printf("  Failed to list collections: %v", err)
			} else {
				log.Printf("  Collections: %v", collections)

				// Check for movies collection
				for _, collName := range collections {
					if collName == "movies" {
						collection := dbInstance.Collection(collName)
						count, err := collection.CountDocuments(context.Background(), nil)
						if err != nil {
							log.Printf("  Failed to count movies: %v", err)
						} else {
							log.Printf("  *** FOUND %d MOVIES in %s.movies ***", count, db.Name)

							if count > 0 {
								// Get a sample document
								var sample map[string]interface{}
								err := collection.FindOne(context.Background(), nil).Decode(&sample)
								if err == nil {
									log.Printf("  Sample movie: %+v", sample)
								}
							}
						}
					}
				}
			}
		}
	}
}
