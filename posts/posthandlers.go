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

		imgList := uniqueIMGsStr(stretchWO)

		c.JSON(200, gin.H{
			"workout": stretchWO,
			"images":  imgList,
		})

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

		imgList := uniqueIMGsWO(workout)

		c.JSON(200, gin.H{
			"workout": workout,
			"images":  imgList,
		})

	}
}

func uniqueIMGsStr(strWO datatypes.StretchWorkout) []string {
	imgMap := map[string]bool{}
	ret := []string{}

	for _, set := range strWO.DynamicSlice {
		for _, rep := range set.RepSlice {
			for _, pos := range rep.Positions {
				for _, img := range pos {
					imgMap[img] = true
				}
			}
		}
	}

	for _, set := range strWO.StaticSlice {
		for _, rep := range set.RepSlice {
			for _, pos := range rep.Positions {
				for _, img := range pos {
					imgMap[img] = true
				}
			}
		}
	}

	for img := range imgMap {
		ret = append(ret, img)
	}

	return ret

}

func uniqueIMGsWO(WO datatypes.Workout) []string {
	imgMap := map[string]bool{}
	ret := []string{}

	for _, set := range WO.DynamicSlice {
		for _, rep := range set.RepSlice {
			for _, pos := range rep.Positions {
				for _, img := range pos {
					imgMap[img] = true
				}
			}
		}
	}

	for _, set := range WO.StaticSlice {
		for _, rep := range set.RepSlice {
			for _, pos := range rep.Positions {
				for _, img := range pos {
					imgMap[img] = true
				}
			}
		}
	}

	for _, round := range WO.Exercises {
		for _, set := range round.SetSlice {
			for _, rep := range set.RepSlice {
				for _, pos := range rep.Positions {
					for _, img := range pos {
						imgMap[img] = true
					}
				}
			}
		}
	}

	for img := range imgMap {
		ret = append(ret, img)
	}

	return ret

}
