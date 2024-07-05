package platform

import (
	"errors"
	"i9-pos/gets"
	"i9-pos/platform/middleware"
	"i9-pos/posts"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const bucketName = "CacheBucket"

type PasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

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

	router.DELETE("/clearcache", clearcache(boltDB))

	return router
}

func temp() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello",
		})
	}
}

func clearcache(db *bbolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req PasswordRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte("$2a$10$cvD77jlhjwTXKLkX.KeI0Ool7Kp5HCjovhJ.mX01P18qxt4R/CbIu"), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("password doesn't match").Error()})
			return
		}

		err = db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				return nil
			}
			return b.ForEach(func(k, v []byte) error {
				return b.Delete(k)
			})
		})

		if err != nil {
			log.Printf("Failed to clear cache: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "error clearing cache",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	}
}
