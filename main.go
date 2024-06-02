package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// HOW TO STORE THE DATA?
// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// STORE WHAT DATA?
// albums slice to seed record album data.
var client *mongo.Client
var albumCollection *mongo.Collection

func main() {
	var err error

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")

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
	albumCollection = client.Database("recordstore").Collection("albums")

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	var albums []album

	cursor, err := albumCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting albums"})
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var album album
		if err := cursor.Decode(&album); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding album"})
			return
		}
		albums = append(albums, album)
	}
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	_, err := albumCollection.InsertOne(context.TODO(), newAlbum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error inserting album"})
		return
	}
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	var album album
	err := albumCollection.FindOne(context.TODO(), bson.D{{Key: "id", Value: id}}).Decode(&album)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
		return

	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
