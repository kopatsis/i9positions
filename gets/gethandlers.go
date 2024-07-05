package gets

import (
	"context"
	"encoding/json"
	"errors"
	"i9-pos/database"
	"i9-pos/datatypes"
	"log"
	"slices"

	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const bucketName = "CacheBucket"

func GetSampleByID(db *mongo.Database, boltDB *bbolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr, exists := c.Params.Get("id")
		if !exists {
			c.JSON(400, gin.H{
				"Error": "Issue with param",
				"Exact": "Unable to get ID from URL parameter",
			})
			return
		}

		sample, err := SampleByID(db, boltDB, idStr)
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

func SampleByID(db *mongo.Database, boltDB *bbolt.DB, id string) (datatypes.Sample, error) {

	samples, err := BoltSamples(db, boltDB)
	if err != nil {
		return datatypes.Sample{}, err
	}

	for _, sample := range samples {
		if sample.ID.Hex() == id {
			return sample, nil
		}
	}

	return datatypes.Sample{}, errors.New("no matches for sample id")
}

func GetSampleByExtID(db *mongo.Database, boltDB *bbolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		typeStr, exists := c.Params.Get("type")
		if !exists || (typeStr != "exercise" && typeStr != "static" && typeStr != "dynamic") {
			c.JSON(400, gin.H{
				"Error": "Issue with param",
				"Exact": "Unable to get ID from URL parameter",
			})
			return
		}

		idStr, exists := c.Params.Get("id")
		if !exists {
			c.JSON(400, gin.H{
				"Error": "Issue with param",
				"Exact": "Unable to get ID from URL parameter",
			})
			return
		}

		sample, err := SampleByExtID(db, boltDB, idStr, typeStr)
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

func SampleByExtID(db *mongo.Database, boltDB *bbolt.DB, id, typeStr string) (datatypes.Sample, error) {

	var formattedType string
	if typeStr == "exercise" {
		formattedType = "Exercise"
	} else if typeStr == "static" {
		formattedType = "Static Stretch"
	} else {
		formattedType = "Dynamic Stretch"
	}

	samples, err := BoltSamples(db, boltDB)
	if err != nil {
		return datatypes.Sample{}, err
	}

	for _, sample := range samples {
		if sample.ExOrStID == id && sample.Type == formattedType {
			return sample, nil
		}
	}

	return datatypes.Sample{}, errors.New("no sample matches provided id")
}

func GetSamples(db *mongo.Database, boltDB *bbolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		if idList, ok := c.GetQueryArray("idList"); ok {
			samples, err := GetSamplesByList(db, boltDB, idList)
			if err != nil {
				c.JSON(400, gin.H{
					"Error": "Issue with querying samples",
					"Exact": err.Error(),
				})
				return
			}

			c.JSON(200, samples)
		} else {
			samples, err := BoltSamples(db, boltDB)
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

func GetSamplesByList(db *mongo.Database, boltDB *bbolt.DB, idList []string) (map[string]datatypes.Sample, error) {
	samples := map[string]datatypes.Sample{}

	uniqueSampleIDs := database.UniqueStrSlice(idList)

	sampleSlice, err := BoltSamples(db, boltDB)
	if err != nil {
		return nil, err
	}

	for _, sample := range sampleSlice {
		if slices.Contains(uniqueSampleIDs, sample.ID.Hex()) {
			samples[sample.ID.Hex()] = sample
		}
	}

	return samples, nil

}

func BoltSamples(database *mongo.Database, boltDB *bbolt.DB) ([]datatypes.Sample, error) {
	var samples []datatypes.Sample

	err := boltDB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		v := b.Get([]byte("Sample"))
		if v != nil {
			err := json.Unmarshal(v, &samples)
			if err == nil {
				return nil
			}
			log.Printf("Failed to unmarshal from bbolt: %v, fetching from MongoDB", err)
		}

		cursor, err := database.Collection("sample").Find(context.Background(), bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(context.Background())

		if err = cursor.All(context.Background(), &samples); err != nil {
			return err
		}

		data, err := json.Marshal(samples)
		if err != nil {
			return err
		}

		err = b.Put([]byte("Sample"), data)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return samples, nil
}
