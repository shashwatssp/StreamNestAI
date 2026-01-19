package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() *mongo.Client {

	err := godotenv.Load(".env.local")

	if err != nil {
		log.Println("WARNING: Unable to Find .env.local file")
	}

	MongoDB := os.Getenv("MONGODB_URI")

	if MongoDB == "" {
		log.Fatal("FATAL: MongoDB URI not set in environment variables")
	}

	log.Printf("INFO: Attempting to connect to MongoDB with URI: %s", MongoDB)

	clientOptions := options.Client().ApplyURI(MongoDB)
	clientOptions.SetConnectTimeout(10 * time.Second)
	clientOptions.SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(clientOptions)

	if err != nil {
		log.Printf("ERROR: Failed to connect to MongoDB: %v", err)
		return nil
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("ERROR: Failed to ping MongoDB: %v", err)
		return nil
	}

	log.Println("SUCCESS: Successfully connected to MongoDB")
	return client
}

var Client *mongo.Client = Connect()

func OpenCollection(collectionName string, client *mongo.Client) *mongo.Collection {

	err := godotenv.Load(".env.local")
	if err != nil {
		log.Println("WARNING: Unable to find .env.local file in OpenCollection")
	}

	databaseName := os.Getenv("DATABASE_NAME")

	log.Printf("INFO: Opening collection '%s' from database '%s'", collectionName, databaseName)

	if databaseName == "" {
		log.Printf("WARNING: DATABASE_NAME not set, using default 'streamnestai'")
		databaseName = "streamnestai"
	}

	collection := client.Database(databaseName).Collection(collectionName)

	if collection == nil {
		log.Printf("ERROR: Failed to get collection '%s' from database '%s'", collectionName, databaseName)
		return nil
	}

	log.Printf("SUCCESS: Successfully opened collection '%s'", collectionName)
	return collection

}
