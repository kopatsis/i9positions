package posts

import (
	"i9-pos/database"
	"i9-pos/datatypes"

	"go.mongodb.org/mongo-driver/mongo"
)

func Workout(db *mongo.Database, resolution string, WOBody datatypes.WorkoutRoute) (datatypes.Workout, error) {
	workout := datatypes.Workout{}

	exerIDRoundList := [9][]string{}
	for i, workoutRound := range WOBody.Exercises {
		exerIDRoundList[i] = workoutRound.ExerciseIDs
	}

	dynamics, statics, imagesets, exercises, err := database.QueryWO(db, WOBody.Statics, WOBody.Dynamics, exerIDRoundList)
	if err != nil {
		return datatypes.Workout{}, nil
	}

	dynamicSets := DynamicSets(dynamics, WOBody.Dynamics, WOBody.StretchTimes, resolution, imagesets)
	workout.DynamicSlice = dynamicSets

	staticSets := StaticSets(statics, WOBody.Statics, WOBody.StretchTimes, resolution, imagesets)
	workout.StaticSlice = staticSets

	workout.DynamicRest = WOBody.StretchTimes.DynamicRest
	workout.DynamicTime = WOBody.StretchTimes.FullRound
	workout.StaticTime = WOBody.StretchTimes.FullRound

	retExers := [9]datatypes.WORound{}
	for i, round := range WOBody.Exercises {
		currentRound := datatypes.WORound{}

		if round.Status == "Regular" {

		} else if round.Status == "Combo" {

		} else {

		}

		retExers[i] = currentRound
	}
	workout.Exercises = retExers

	return workout, nil
}
