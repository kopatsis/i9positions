package platform

import (
	"i9-pos/gets"
	"i9-pos/platform/middleware"
	"i9-pos/posts"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func New(database *mongo.Database) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.JWTAuthMiddleware())

	router.GET("/", temp(database))

	router.GET("/samples", gets.GetSamples(database))
	router.GET("/samples/:id", gets.GetSampleByID(database))

	router.POST("/workouts/stretch/:res", posts.PostStretchWorkout(database))
	router.POST("/workouts/:res", posts.PostWorkout(database))

	return router
}

func temp(database *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello",
		})
	}
}
