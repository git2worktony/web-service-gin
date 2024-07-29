package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// profile represents data about a user's profile.
type profile struct {
	User                  string `json:"user"`
	Address               string `json:"address"`
	ResponsibleIndividual string `json:"responsible_individual"`
	ContactNumber         string `json:"contact_number"`
}

// profileCollection to handle profile data.
var client *mongo.Client
var profileCollection *mongo.Collection

func main() {
	var err error

	// Use the MONGO_URI environment variable
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	// Get a handle for your collection
	profileCollection = client.Database("marketplace").Collection("profiles")

	router := gin.Default()
	router.GET("/profiles", getProfiles)
	router.GET("/profiles/:user", getProfileByUser)
	router.POST("/profiles", postProfiles)

	router.Run(":8080")
}

// getProfiles responds with the list of all profiles as JSON.
func getProfiles(c *gin.Context) {
	var profiles []profile

	cursor, err := profileCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting profiles"})
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var profile profile
		if err := cursor.Decode(&profile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding profile"})
			return
		}
		profiles = append(profiles, profile)
	}
	c.IndentedJSON(http.StatusOK, profiles)
}

// postProfiles adds a profile from JSON received in the request body.
func postProfiles(c *gin.Context) {
	var newProfile profile

	// Call BindJSON to bind the received JSON to
	// newProfile.
	if err := c.BindJSON(&newProfile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	_, err := profileCollection.InsertOne(context.TODO(), newProfile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error inserting profile"})
		return
	}
	c.IndentedJSON(http.StatusCreated, newProfile)
}

// getProfileByUser locates the profile whose User value matches the user
// parameter sent by the client, then returns that profile as a response.
func getProfileByUser(c *gin.Context) {
	user := c.Param("user")

	var profile profile
	err := profileCollection.FindOne(context.TODO(), bson.D{{Key: "user", Value: user}}).Decode(&profile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "profile not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, profile)
}
