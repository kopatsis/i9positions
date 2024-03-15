package posts

import (
	"i9-pos/datatypes"
	"slices"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func PostStretchWorkout(database *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		resolution, exists := c.Params.Get("res")
		if !exists || !slices.Contains([]string{"Low", "Medium", "High", "Original"}, resolution) {
			resolution = "High"
		}

		var strWOBody datatypes.StretchWorkoutRoute
		if err := c.ShouldBindJSON(&strWOBody); err != nil {
			c.JSON(400, gin.H{
				"Error": "Issue with body binding",
				"Exact": err.Error(),
			})
			return
		}

		stretchWO, err := StretchWorkout(database, resolution, strWOBody)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Issue with stretch WO creation",
				"Exact": err.Error(),
			})
			return
		}

		c.JSON(200, stretchWO)

	}
}

func PostWorkout(database *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		resolution, exists := c.Params.Get("res")
		if !exists || !slices.Contains([]string{"Low", "Medium", "High", "Original"}, resolution) {
			resolution = "High"
		}

		var WOBody datatypes.WorkoutRoute
		if err := c.ShouldBindJSON(&WOBody); err != nil {
			c.JSON(400, gin.H{
				"Error": "Issue with body binding",
				"Exact": err.Error(),
			})
			return
		}

		workout, err := Workout(database, resolution, WOBody)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Issue with WO creation",
				"Exact": err.Error(),
			})
			return
		}

		c.JSON(200, workout)

	}
}
