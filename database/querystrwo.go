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

func QueryStretchWO(database *mongo.Database, boltDB *bbolt.DB, statics, dynamics []string) (map[string]datatypes.DynamicStr, map[string]datatypes.StaticStr, error) {
	var wg sync.WaitGroup

	errChan := make(chan error, 2)
	var errGroup *multierror.Error
	dynamicStr, staticStr := map[string]datatypes.DynamicStr{}, map[string]datatypes.StaticStr{}
	// var imageSets map[string]datatypes.ImageSet

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
		return nil, nil, errGroup
	}

	return dynamicStr, staticStr, nil
}

func GetDynamics(database *mongo.Database, boltDB *bbolt.DB, dynamics []string) (map[string]datatypes.DynamicStr, error) {

	dynamicStr := map[string]datatypes.DynamicStr{}

	var dynamicList []datatypes.DynamicStr

	err := boltDB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		v := b.Get([]byte("Dynamic"))
		if v != nil {
			err := json.Unmarshal(v, &dynamicList)
			if err == nil {
				return nil
			}
			log.Printf("Failed to unmarshal from bbolt: %v, fetching from MongoDB", err)
		}

		cursor, err := database.Collection("dynamicstretch").Find(context.Background(), bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(context.Background())

		if err = cursor.All(context.Background(), &dynamicList); err != nil {
			return err
		}

		data, err := json.Marshal(dynamicList)
		if err != nil {
			return err
		}

		err = b.Put([]byte("Dynamic"), data)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, dynamic := range dynamicList {
		if slices.Contains(dynamics, dynamic.BackendID) {
			dynamicStr[dynamic.BackendID] = dynamic
		}
	}

	return dynamicStr, nil
}

func GetStatics(database *mongo.Database, boltDB *bbolt.DB, statics []string) (map[string]datatypes.StaticStr, error) {
	staticStr := map[string]datatypes.StaticStr{}

	var staticList []datatypes.StaticStr

	err := boltDB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		v := b.Get([]byte("Static"))
		if v != nil {
			err := json.Unmarshal(v, &staticList)
			if err == nil {
				return nil
			}
			log.Printf("Failed to unmarshal from bbolt: %v, fetching from MongoDB", err)
		}

		cursor, err := database.Collection("staticstretch").Find(context.Background(), bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(context.Background())

		if err = cursor.All(context.Background(), &staticList); err != nil {
			return err
		}

		data, err := json.Marshal(staticList)
		if err != nil {
			return err
		}

		err = b.Put([]byte("Static"), data)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, static := range staticList {
		if slices.Contains(statics, static.BackendID) {
			staticStr[static.BackendID] = static
		}
	}

	return staticStr, nil
}

func UniqueStrSlice(sl []string) []string {

	ret := []string{}
	contains := map[string]bool{}
	for _, s := range sl {
		if _, ok := contains[s]; !ok {
			ret = append(ret, s)
			contains[s] = true
		}
	}
	return ret
}
