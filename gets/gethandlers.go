package gets

import (
	"context"
	"i9-pos/datatypes"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetSampleByID(database *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr, exists := c.Params.Get("id")
		if !exists {
			c.JSON(400, gin.H{
				"Error": "Issue with param",
				"Exact": "Unable to get ID from URL parameter",
			})
			return
		}

		sample, err := SampleByID(database, idStr)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Issue with querying sample",
				"Exact": err.Error(),
			})
			return
		}

		c.JSON(200, sample)

	}
}

func SampleByID(database *mongo.Database, id string) (datatypes.Sample, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return datatypes.Sample{}, err
	}

	collection := database.Collection("sample")

	var result datatypes.Sample
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return datatypes.Sample{}, err
	}

	return result, nil
}

func GetSamples(database *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		samples, err := AllSamples(database)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Issue with querying samples",
				"Exact": err.Error(),
			})
			return
		}

		c.JSON(200, samples)

	}
}

func AllSamples(database *mongo.Database) ([]datatypes.Sample, error) {

	collection := database.Collection("sample")

	var samples []datatypes.Sample
	cur, err := collection.Find(context.TODO(), bson.D{{}}, options.Find())
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var sample datatypes.Sample
		err := cur.Decode(&sample)
		if err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return samples, nil
}
