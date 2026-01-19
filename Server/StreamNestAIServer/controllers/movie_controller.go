package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/joho/godotenv"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/database"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/models"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/gin-gonic/gin"
)

var validate = validator.New()

func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("INFO: GetMovies endpoint called")
		
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movies []models.Movie

		log.Println("INFO: Opening movies collection")
		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)

		if movieCollection == nil {
			log.Println("ERROR: Failed to open movies collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Open Movies Collection"})
			return
		}

		log.Println("INFO: Executing Find query on movies collection")
		cursor, err := movieCollection.Find(ctx, bson.M{})

		if err != nil {
			log.Printf("ERROR: Failed to fetch movies from database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Fetch Movies"})
			return
		}

		defer cursor.Close(ctx)

		log.Println("INFO: Decoding cursor results into movies slice")
		if err = cursor.All(ctx, &movies); err != nil {
			log.Printf("ERROR: Failed to decode movies: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Decode Movies"})
			return
		}

		log.Printf("INFO: Successfully fetched %d movies from database", len(movies))
		
		if len(movies) == 0 {
			log.Println("WARNING: No movies found in database - collection might be empty")
			// Ensure we return an empty array instead of null
			c.JSON(http.StatusOK, []models.Movie{})
			return
		} else {
			log.Printf("DEBUG: First movie title: %s", movies[0].Title)
		}

		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		log.Printf("INFO: GetMovie endpoint called at %s", startTime.Format(time.RFC3339))
		
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieID := c.Param("imdb_id")
		log.Printf("INFO: Looking for movie with imdb_id: %s", movieID)

		if movieID == "" {
			log.Printf("ERROR: Movie ID is empty in request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}

		log.Printf("INFO: Opening collection 'movies' from database 'StreamNest'")
		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		
		if movieCollection == nil {
			log.Printf("ERROR: Failed to open movies collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Open Movies Collection"})
			return
		}
		
		log.Printf("SUCCESS: Successfully opened collection 'movies'")
		log.Printf("INFO: Executing FindOne query for imdb_id: %s", movieID)
		
		var movie models.Movie
		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.Printf("WARNING: Movie with imdb_id '%s' not found in database", movieID)
				c.JSON(http.StatusNotFound, gin.H{"error": "Movie Not Found"})
			} else {
				log.Printf("ERROR: Database error while fetching movie: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			}
			return
		}

		duration := time.Since(startTime)
		log.Printf("INFO: Successfully found movie '%s' (imdb_id: %s) in %v", movie.Title, movieID, duration)
		log.Printf("DEBUG: Movie details - Title: %s, Genres: %v, Ranking: %v",
			movie.Title, movie.Genre, movie.Ranking)

		c.JSON(http.StatusOK, movie)
	}
}

func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movie models.Movie

		err := c.ShouldBindJSON(&movie)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Some Error Occured"})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Failed Details ", "details": err.Error()})
			return
		}
		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := movieCollection.InsertOne(ctx, movie)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Add Movie"})
			return
		}

		c.JSON(http.StatusCreated, result)

	}
}

func AdminReviewUpdate(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		role, err := utils.GetRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
			return
		}

		if role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User must be part of the ADMIN role"})
			return
		}

		movieId := c.Param("imdb_id")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie Id required"})
			return
		}
		var req struct {
			AdminReview string `json:"admin_review"`
		}
		var resp struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		sentiment, rankVal, err := GetReviewRanking(req.AdminReview, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting review ranking"})
			return
		}

		filter := bson.D{{Key: "imdb_id", Value: movieId}}

		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_value": rankVal,
					"ranking_name":  sentiment,
				},
			},
		}
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)

		result, err := movieCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movie"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		resp.RankingName = sentiment
		resp.AdminReview = req.AdminReview

		c.JSON(http.StatusOK, resp)

	}
}
func GetReviewRanking(admin_review string, client *mongo.Client, c *gin.Context) (string, int, error) {
	log.Println("Starting GetReviewRanking")

	rankings, err := GetRankings(client, c)
	if err != nil {
		log.Println("Error fetching rankings:", err)
		return "", 0, err
	}
	log.Println("Rankings fetched:", rankings)

	sentimentDelimited := ""
	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited += ranking.RankingName + ","
		}
	}
	sentimentDelimited = strings.Trim(sentimentDelimited, ",")
	log.Println("Sentiment options:", sentimentDelimited)

	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// You can use either DEEPSEEK_API_KEY or keep OPENAI_API_KEY
	OpenAiApiKey := os.Getenv("OPENAI_API_KEY")
	log.Println("API Key loaded:", OpenAiApiKey != "")

	if OpenAiApiKey == "" {
		log.Println("API Key is empty - using fallback sentiment analysis")
		
		// Simple fallback sentiment analysis for local development
		reviewLower := strings.ToLower(admin_review)
		var sentiment string
		var rankVal int
		
		// Simple keyword-based sentiment analysis
		if strings.Contains(reviewLower, "excellent") || strings.Contains(reviewLower, "amazing") || strings.Contains(reviewLower, "fantastic") || strings.Contains(reviewLower, "love") || strings.Contains(reviewLower, "perfect") {
			sentiment = "Excellent"
			rankVal = 1
		} else if strings.Contains(reviewLower, "good") || strings.Contains(reviewLower, "great") || strings.Contains(reviewLower, "nice") || strings.Contains(reviewLower, "enjoyed") {
			sentiment = "Good"
			rankVal = 2
		} else if strings.Contains(reviewLower, "okay") || strings.Contains(reviewLower, "fine") || strings.Contains(reviewLower, "decent") || strings.Contains(reviewLower, "average") {
			sentiment = "Okay"
			rankVal = 3
		} else if strings.Contains(reviewLower, "bad") || strings.Contains(reviewLower, "poor") || strings.Contains(reviewLower, "disappoint") || strings.Contains(reviewLower, "didn't like") {
			sentiment = "Bad"
			rankVal = 4
		} else if strings.Contains(reviewLower, "terrible") || strings.Contains(reviewLower, "awful") || strings.Contains(reviewLower, "hate") || strings.Contains(reviewLower, "worst") {
			sentiment = "Terrible"
			rankVal = 5
		} else {
			// Default to "Okay" if no clear sentiment detected
			sentiment = "Okay"
			rankVal = 3
		}
		
		log.Printf("Fallback sentiment analysis result: %s (rank: %d)", sentiment, rankVal)
		return sentiment, rankVal, nil
	}

	// Initialize with DeepSeek base URL and model
	llm, err := openai.New(
		openai.WithToken(OpenAiApiKey),
		openai.WithBaseURL("https://api.deepseek.com"),
		openai.WithModel("deepseek-chat"), // or "deepseek-reasoner" for R1
	)
	if err != nil {
		log.Println("Error creating DeepSeek client:", err)
		return "", 0, err
	}

	base_prompt_template := os.Getenv("BASE_PROMPT_TEMPLATE")
	log.Println("Base prompt template:", base_prompt_template)

	if base_prompt_template == "" {
		log.Println("BASE_PROMPT_TEMPLATE is empty - using fallback")
		// Use a simple fallback prompt template
		base_prompt_template = fmt.Sprintf("Analyze the sentiment of this movie review and respond with exactly one of these ratings: %s. Review: ", sentimentDelimited)
	}

	base_prompt := strings.Replace(base_prompt_template, "{rankings}", sentimentDelimited, 1)
	fullPrompt := base_prompt + admin_review
	log.Println("Full prompt sent to DeepSeek:", fullPrompt)

	response, err := llm.Call(c, fullPrompt)
	if err != nil {
		log.Println("DeepSeek call error:", err)
		log.Println("Falling back to local sentiment analysis due to API failure")
		
		// Fallback sentiment analysis when API fails
		reviewLower := strings.ToLower(admin_review)
		var sentiment string
		var rankVal int
		
		// Simple keyword-based sentiment analysis
		if strings.Contains(reviewLower, "excellent") || strings.Contains(reviewLower, "amazing") || strings.Contains(reviewLower, "fantastic") || strings.Contains(reviewLower, "love") || strings.Contains(reviewLower, "perfect") || strings.Contains(reviewLower, "sublime") {
			sentiment = "Excellent"
			rankVal = 1
		} else if strings.Contains(reviewLower, "good") || strings.Contains(reviewLower, "great") || strings.Contains(reviewLower, "nice") || strings.Contains(reviewLower, "enjoyed") || strings.Contains(reviewLower, "wonderful") {
			sentiment = "Good"
			rankVal = 2
		} else if strings.Contains(reviewLower, "okay") || strings.Contains(reviewLower, "fine") || strings.Contains(reviewLower, "decent") || strings.Contains(reviewLower, "average") {
			sentiment = "Okay"
			rankVal = 3
		} else if strings.Contains(reviewLower, "bad") || strings.Contains(reviewLower, "poor") || strings.Contains(reviewLower, "disappoint") || strings.Contains(reviewLower, "didn't like") {
			sentiment = "Bad"
			rankVal = 4
		} else if strings.Contains(reviewLower, "terrible") || strings.Contains(reviewLower, "awful") || strings.Contains(reviewLower, "hate") || strings.Contains(reviewLower, "worst") {
			sentiment = "Terrible"
			rankVal = 5
		} else {
			// Default to "Okay" if no clear sentiment detected
			sentiment = "Okay"
			rankVal = 3
		}
		
		log.Printf("Fallback sentiment analysis result: %s (rank: %d)", sentiment, rankVal)
		return sentiment, rankVal, nil
	}
	log.Println("DeepSeek response:", response)

	rankVal := 0
	for _, ranking := range rankings {
		if ranking.RankingName == response {
			rankVal = ranking.RankingValue
			break
		}
	}
	log.Println("Rank value matched:", rankVal)

	return response, rankVal, nil
}

func GetRankings(client *mongo.Client, c *gin.Context) ([]models.Ranking, error) {
	var rankings []models.Ranking

	var ctx, cancel = context.WithTimeout(c, 100*time.Second)
	defer cancel()

	var rankingCollection *mongo.Collection = database.OpenCollection("rankings", client)

	cursor, err := rankingCollection.Find(ctx, bson.D{})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil

}

func GetRecommendedMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := utils.GetUserIdFromContext(c)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User Id not found in context"})
		}

		favourite_genres, err := GetUsersFavouriteGenres(userId, client, c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}
		var recommendedMovieLimitVal int64 = 5

		recommendedMovieLimitStr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")

		if recommendedMovieLimitStr != "" {
			recommendedMovieLimitVal, _ = strconv.ParseInt(recommendedMovieLimitStr, 10, 64)
		}

		findOptions := options.Find()

		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}})

		findOptions.SetLimit(recommendedMovieLimitVal)

		filter := bson.D{
			{Key: "genre.genre_name", Value: bson.D{
				{Key: "$in", Value: favourite_genres},
			}},
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)

		cursor, err := movieCollection.Find(ctx, filter, findOptions)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies"})
			return
		}
		defer cursor.Close(ctx)

		var recommendedMovies []models.Movie

		if err := cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, recommendedMovies)
	}
}

func GetUsersFavouriteGenres(userId string, client *mongo.Client, c *gin.Context) ([]string, error) {

	var ctx, cancel = context.WithTimeout(c, 100*time.Second)
	defer cancel()

	filter := bson.D{{Key: "user_id", Value: userId}}

	projection := bson.M{
		"favourite_genres.genre_name": 1,
		"_id":                         0,
	}

	opts := options.FindOne().SetProjection(projection)
	var result bson.M

	var userCollection *mongo.Collection = database.OpenCollection("users", client)
	err := userCollection.FindOne(ctx, filter, opts).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
	}

	favGenresArray, ok := result["favourite_genres"].(bson.A)

	if !ok {
		return []string{}, errors.New("unable to retrieve favourite genres for user")
	}

	var genreNames []string

	for _, item := range favGenresArray {
		if genreMap, ok := item.(bson.D); ok {
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}
	}

	return genreNames, nil

}

func GetGenres(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var genreCollection *mongo.Collection = database.OpenCollection("genres", client)

		cursor, err := genreCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movie genres"})
			return
		}
		defer cursor.Close(ctx)

		var genres []models.Genre
		if err := cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, genres)

	}
}

func NaturalLanguageMovieSearch(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		log.Printf("INFO: NaturalLanguageMovieSearch endpoint called at %s", startTime.Format(time.RFC3339))
		
		query := c.Query("q")
		log.Printf("INFO: Search query received: '%s'", query)

		if query == "" {
			log.Printf("ERROR: Empty search query provided")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		// Using existing DeepSeek setup to convert NL to MongoDB filter
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		OpenAiApiKey := os.Getenv("OPENAI_API_KEY")
		log.Printf("INFO: OpenAI API Key loaded: %t", OpenAiApiKey != "")

		llm, err := openai.New(
			openai.WithToken(OpenAiApiKey),
			openai.WithBaseURL("https://api.deepseek.com"),
			openai.WithModel("deepseek-chat"),
		)

		if err != nil {
			log.Printf("ERROR: Failed to initialize DeepSeek client: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize AI"})
			return
		}

		// Prompt to convert natural language to MongoDB filter
		prompt := fmt.Sprintf(`Convert this natural language movie search into MongoDB filter JSON.
User query: "%s"

Available fields:
- title (string)
- genre.genre_name (string)
- ranking.ranking_value (int, 1-10, lower is better)
- year (int)

Examples:
"action movies from 2020" → {"genre.genre_name": "Action", "year": 2020}
"highly rated sci-fi" → {"genre.genre_name": "Sci-Fi", "ranking.ranking_value": {"$lte": 3}}

Return ONLY valid MongoDB filter JSON, no explanation:`, query)

		log.Printf("INFO: Sending query to DeepSeek for processing: '%s'", query)
		filterJSON, err := llm.Call(c, prompt)
		if err != nil {
			log.Printf("ERROR: DeepSeek API call failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI processing failed"})
			return
		}

		log.Printf("INFO: DeepSeek response: '%s'", filterJSON)

		// Parse the AI-generated filter
		var filter bson.M
		if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
			log.Printf("WARNING: Failed to parse AI filter JSON: %v, using fallback", err)
			// Fallback: search by title if AI output is invalid
			filter = bson.M{"title": bson.M{"$regex": query, "$options": "i"}}
			log.Printf("INFO: Using fallback filter: %v", filter)
		} else {
			log.Printf("INFO: Successfully parsed AI filter: %v", filter)
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		log.Printf("INFO: Opening collection 'movies' from database 'StreamNest'")
		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		
		if movieCollection == nil {
			log.Printf("ERROR: Failed to open movies collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Open Movies Collection"})
			return
		}

		// Limit results to 20
		findOptions := options.Find().SetLimit(20)
		log.Printf("INFO: Executing search with filter: %v", filter)

		cursor, err := movieCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("ERROR: Database query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err := cursor.All(ctx, &movies); err != nil {
			log.Printf("ERROR: Failed to decode search results: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		duration := time.Since(startTime)
		log.Printf("INFO: Natural language search completed in %v, found %d results", duration, len(movies))
		
		if len(movies) > 0 {
			log.Printf("DEBUG: First result: %s (imdb_id: %s)", movies[0].Title, movies[0].ImdbID)
		}

		c.JSON(http.StatusOK, movies)
	}
}
