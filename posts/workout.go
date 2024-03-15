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

		currentRound.SetCount = round.Times.Sets
		currentRound.FullTime = round.Times.FullRound
		currentRound.RestPerRound = round.Times.RestPerRound
		currentRound.RestPerSet = round.Times.RestPerSet

		if round.Status == "Regular" {
			currentRound.SetSlice, currentRound.SetSequence = RegularRound(exercises)
		} else if round.Status == "Combo" {
			currentRound.SetSlice, currentRound.SetSequence = RegularRound(exercises) // remove
		} else {
			currentRound.SetSlice, currentRound.SetSequence = RegularRound(exercises) // remove
		}

		retExers[i] = currentRound
	}
	workout.Exercises = retExers

	workout.BackendID = WOBody.ID.Hex()

	return workout, nil
}

func RegularRound(exercises map[string]datatypes.Exercise) ([]datatypes.Set, []int) {
	return []datatypes.Set{}, []int{}
}
