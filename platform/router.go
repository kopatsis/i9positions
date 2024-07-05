package platform

import (
	"i9-pos/gets"
	"i9-pos/platform/middleware"
	"i9-pos/posts"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/mongo"
)

func New(database *mongo.Database, firebase *firebase.App, boltDB *bbolt.DB) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.JWTAuthMiddleware(firebase))

	router.GET("/", temp())

	router.GET("/samples", gets.GetSamples(database, boltDB))
	router.GET("/samples/:id", gets.GetSampleByID(database, boltDB))
	router.GET("/samples/ext/:type/:id", gets.GetSampleByExtID(database, boltDB))

	router.POST("/workouts/stretch", posts.PostStretchWorkout(database, boltDB))
	router.POST("/workouts", posts.PostWorkout(database, boltDB))

	return router
}

func temp() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello",
		})
	}
}
