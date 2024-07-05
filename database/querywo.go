package database

import (
	"context"
	"encoding/json"
	"i9-pos/datatypes"
	"log"
	"slices"
	"sync"

	"github.com/hashicorp/go-multierror"
	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const bucketName = "CacheBucket"

func QueryWO(database *mongo.Database, boltDB *bbolt.DB, noMax bool, statics, dynamics []string, exercises [9][]string) (map[string]datatypes.DynamicStr, map[string]datatypes.StaticStr, map[string]datatypes.Exercise, datatypes.TransitionMatrix, error) {
	var wg sync.WaitGroup

	errChan := make(chan error, 4)
	var errGroup *multierror.Error
	dynamicStr, staticStr, exerciseMap, matrix := map[string]datatypes.DynamicStr{}, map[string]datatypes.StaticStr{}, map[string]datatypes.Exercise{}, datatypes.TransitionMatrix{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		dynamicStr, err = GetDynamics(database, boltDB, dynamics)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		staticStr, err = GetStatics(database, boltDB, statics)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		exerciseMap, err = GetExercises(database, boltDB, exercises)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		matrix, err = GetTransitionMatrix(database, boltDB)
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
		return nil, nil, nil, matrix, errGroup
	}

	if noMax {
		for id, ex := range exerciseMap {
			ex.MaxSecs = 999
			exerciseMap[id] = ex
		}
	}

	return dynamicStr, staticStr, exerciseMap, matrix, nil
}

func GetExercises(database *mongo.Database, boltDB *bbolt.DB, exercises [9][]string) (map[string]datatypes.Exercise, error) {
	exerciseMap := map[string]datatypes.Exercise{}

	sumIdList := []string{}
	for _, idlist := range exercises {
		sumIdList = append(sumIdList, idlist...)
	}

	uniqueIDList := UniqueStrSlice(sumIdList)

	var exerciseList []datatypes.Exercise

	err := boltDB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		v := b.Get([]byte("Exercise"))
		if v != nil {
			err := json.Unmarshal(v, &exerciseList)
			if err == nil {
				return nil
			}
			log.Printf("Failed to unmarshal from bbolt: %v, fetching from MongoDB", err)
		}

		cursor, err := database.Collection("exercise").Find(context.Background(), bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(context.Background())

		if err = cursor.All(context.Background(), &exerciseList); err != nil {
			return err
		}

		data, err := json.Marshal(exerciseList)
		if err != nil {
			return err
		}

		err = b.Put([]byte("Exercise"), data)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, exer := range exerciseList {
		if slices.Contains(uniqueIDList, exer.BackendID) {
			exerciseMap[exer.BackendID] = exer
		}
	}

	return exerciseMap, nil
}

func GetTransitionMatrix(database *mongo.Database, boltDB *bbolt.DB) (datatypes.TransitionMatrix, error) {

	var matrix datatypes.TransitionMatrix

	err := boltDB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		v := b.Get([]byte("Transition"))
		if v != nil {
			err := json.Unmarshal(v, &matrix)
			if err == nil {
				return nil
			}
			log.Printf("Failed to unmarshal from bbolt: %v, fetching from MongoDB", err)
		}

		err = database.Collection("transition").FindOne(context.Background(), bson.D{}).Decode(&matrix)
		if err != nil {
			return err
		}

		data, err := json.Marshal(matrix)
		if err != nil {
			return err
		}

		err = b.Put([]byte("Transition"), data)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return datatypes.TransitionMatrix{}, err
	}

	return matrix, nil
}
