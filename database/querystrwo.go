package database

import (
	"context"
	"i9-pos/datatypes"
	"sync"

	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func QueryStretchWO(database *mongo.Database, statics, dynamics []string) (map[string]datatypes.DynamicStr, map[string]datatypes.StaticStr, map[string]datatypes.ImageSet, error) {
	var wg sync.WaitGroup

	errChan := make(chan error, 2)
	var errGroup *multierror.Error
	dynamicStr, staticStr := map[string]datatypes.DynamicStr{}, map[string]datatypes.StaticStr{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		dynamicStr, err = GetDynamics(database, dynamics)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		staticStr, err = GetStatics(database, statics)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	hasErr := false
	for err := range errChan {
		if err != nil {
			errGroup = multierror.Append(errGroup, err)
			hasErr = true
		}
	}

	if hasErr {
		return nil, nil, nil, errGroup
	}

	imageSets, err := GetImageSets(database, dynamicStr, staticStr)
	if err != nil {
		return nil, nil, nil, err
	}

	return dynamicStr, staticStr, imageSets, nil
}

func GetDynamics(database *mongo.Database, dynamics []string) (map[string]datatypes.DynamicStr, error) {
	dynamicStr := map[string]datatypes.DynamicStr{}

	collection := database.Collection("dynamicstretch")
	filter := bson.M{"backendID": bson.M{"$in": UniqueStrSlice(dynamics)}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.Background()) {
		var result datatypes.DynamicStr
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}

		dynamicStr[result.BackendID] = result
	}

	return dynamicStr, nil
}

func GetStatics(database *mongo.Database, statics []string) (map[string]datatypes.StaticStr, error) {
	staticStr := map[string]datatypes.StaticStr{}

	collection := database.Collection("staticstretch")
	filter := bson.M{"backendID": bson.M{"$in": UniqueStrSlice(statics)}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.Background()) {
		var result datatypes.StaticStr
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}

		staticStr[result.BackendID] = result
	}

	return staticStr, nil
}

func GetImageSets(database *mongo.Database, dynamicStr map[string]datatypes.DynamicStr, staticStr map[string]datatypes.StaticStr) (map[string]datatypes.ImageSet, error) {
	imageSets := map[string]datatypes.ImageSet{}

	allImageSets := []string{}
	for _, dynamic := range dynamicStr {
		for _, position := range dynamic.PositionSlice1 {
			allImageSets = append(allImageSets, position.ImageSetID)
		}
		if len(dynamic.PositionSlice2) > 0 {
			for _, position := range dynamic.PositionSlice2 {
				allImageSets = append(allImageSets, position.ImageSetID)
			}
		}
	}

	for _, static := range staticStr {
		allImageSets = append(allImageSets, static.ImageSetID1)
		if static.ImageSetID2 != "" {
			allImageSets = append(allImageSets, static.ImageSetID2)
		}
	}

	uniqueImageSets := UniqueStrSlice(allImageSets)

	imageSetIDPrims := []primitive.ObjectID{}

	for _, id := range uniqueImageSets {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		imageSetIDPrims = append(imageSetIDPrims, objID)
	}

	collection := database.Collection("imageset")

	filter := bson.M{"_id": bson.M{"$in": imageSetIDPrims}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.Background()) {
		var result datatypes.ImageSet
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}

		imageSets[result.ID.Hex()] = result
	}

	return imageSets, nil
}

func GetSamples(database *mongo.Database, dynamicStr map[string]datatypes.DynamicStr, staticStr map[string]datatypes.StaticStr) (map[string]datatypes.Sample, error) {
	samples := map[string]datatypes.Sample{}

	sampleIDs := []string{}
	for _, dynamic := range dynamicStr {
		sampleIDs = append(sampleIDs, dynamic.SampleID)
	}

	for _, static := range staticStr {
		sampleIDs = append(sampleIDs, static.SampleID)
	}

	uniqueSampleIDs := UniqueStrSlice(sampleIDs)

	sampleIDPrims := []primitive.ObjectID{}

	for _, id := range uniqueSampleIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		sampleIDPrims = append(sampleIDPrims, objID)
	}

	collection := database.Collection("sample")

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

func UniqueStrSlice(sl []string) []string {
	ret := []string{}
	contains := map[string]bool{}
	for _, s := range sl {
		if _, ok := contains[s]; !ok {
			ret = append(ret, s)
		}
	}
	return ret
}
