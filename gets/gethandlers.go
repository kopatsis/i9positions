package gets

import (
	"context"
	"i9-pos/database"
	"i9-pos/datatypes"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetSampleByID(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr, exists := c.Params.Get("id")
		if !exists {
			c.JSON(400, gin.H{
				"Error": "Issue with param",
				"Exact": "Unable to get ID from URL parameter",
			})
			return
		}

		sample, err := SampleByID(db, idStr)
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

func SampleByID(db *mongo.Database, id string) (datatypes.Sample, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return datatypes.Sample{}, err
	}

	collection := db.Collection("sample")

	var result datatypes.Sample
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return datatypes.Sample{}, err
	}

	return result, nil
}

func GetSamples(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		if idList, ok := c.GetQueryArray("idList"); ok {
			samples, err := GetSamplesByList(db, idList)
			if err != nil {
				c.JSON(400, gin.H{
					"Error": "Issue with querying samples",
					"Exact": err.Error(),
				})
				return
			}

			c.JSON(200, samples)
		} else {
			samples, err := AllSamples(db)
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
}

func AllSamples(db *mongo.Database) ([]datatypes.Sample, error) {

	collection := db.Collection("sample")

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

func GetSamplesByList(db *mongo.Database, idList []string) (map[string]datatypes.Sample, error) {
	samples := map[string]datatypes.Sample{}

	uniqueSampleIDs := database.UniqueStrSlice(idList)

	sampleIDPrims := []primitive.ObjectID{}

	for _, id := range uniqueSampleIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		sampleIDPrims = append(sampleIDPrims, objID)
	}

	collection := db.Collection("sample")

	filter := bson.M{"_id": bson.M{"$in": sampleIDPrims}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.Background()) {
		var result datatypes.Sample
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}

		samples[result.ID.Hex()] = result
	}

	return samples, nil
}
