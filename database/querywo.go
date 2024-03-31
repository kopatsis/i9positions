package database

import (
	"context"
	"fmt"
	"i9-pos/datatypes"
	"sync"

	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func QueryWO(database *mongo.Database, statics, dynamics []string, exercises [9][]string) (map[string]datatypes.DynamicStr, map[string]datatypes.StaticStr, map[string]datatypes.ImageSet, map[string]datatypes.Exercise, datatypes.TransitionMatrix, error) {
	var wg sync.WaitGroup

	errChan := make(chan error, 5)
	var errGroup *multierror.Error
	dynamicStr, staticStr, exerciseMap, matrix := map[string]datatypes.DynamicStr{}, map[string]datatypes.StaticStr{}, map[string]datatypes.Exercise{}, datatypes.TransitionMatrix{}
	var imageSets map[string]datatypes.ImageSet

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		dynamicStr, err = GetDynamics(database, dynamics)
		if err != nil {
			fmt.Println("1")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		staticStr, err = GetStatics(database, statics)
		if err != nil {
			fmt.Println("2")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		exerciseMap, err = GetExercises(database, exercises)
		if err != nil {
			fmt.Println("3")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		matrix, err = GetTransitionMatrix(database)
		if err != nil {
			fmt.Println("4")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		imageSets, err = GetAllImageSets(database)
		if err != nil {
			fmt.Println("5")
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
		return nil, nil, nil, nil, matrix, errGroup
	}

	// imageSets, err := GetImageSetsWO(database, dynamicStr, staticStr, exerciseMap)
	// if err != nil {
	// 	return nil, nil, nil, nil, matrix, err
	// }

	return dynamicStr, staticStr, imageSets, exerciseMap, matrix, nil
}

func GetExercises(database *mongo.Database, exercises [9][]string) (map[string]datatypes.Exercise, error) {
	exerciseMap := map[string]datatypes.Exercise{}

	sumIdList := []string{}
	for _, idlist := range exercises {
		sumIdList = append(sumIdList, idlist...)
	}

	collection := database.Collection("exercise")
	filter := bson.M{"backendID": bson.M{"$in": UniqueStrSlice(sumIdList)}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.Background()) {
		var result datatypes.Exercise
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}

		exerciseMap[result.BackendID] = result
	}

	return exerciseMap, nil
}

func GetTransitionMatrix(database *mongo.Database) (datatypes.TransitionMatrix, error) {
	matrix := datatypes.TransitionMatrix{}

	collection := database.Collection("transition")
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&matrix)
	if err != nil {
		return matrix, err
	}

	return matrix, nil
}

func GetImageSetsWO(database *mongo.Database, dynamicStr map[string]datatypes.DynamicStr, staticStr map[string]datatypes.StaticStr, exerciseMap map[string]datatypes.Exercise) (map[string]datatypes.ImageSet, error) {
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

	for _, exer := range exerciseMap {
		for _, position := range exer.PositionSlice1 {
			allImageSets = append(allImageSets, position.ImageSetID)
		}
		if len(exer.PositionSlice2) > 0 {
			for _, position := range exer.PositionSlice2 {
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
