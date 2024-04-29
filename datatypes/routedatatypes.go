package datatypes

import "go.mongodb.org/mongo-driver/bson/primitive"

type StretchWorkoutRoute struct {
	Dynamics     []string
	Statics      []string
	StretchTimes StretchTimes
	ID           primitive.ObjectID
}

type StretchTimes struct {
	DynamicPerSet []float32
	StaticPerSet  []float32
	DynamicSets   int
	StaticSets    int
	DynamicRest   float32
	FullRound     float32
}

type WorkoutRoute struct {
	Dynamics     []string
	Statics      []string
	StretchTimes StretchTimes
	ID           primitive.ObjectID
	Difficulty   int
	Exercises    [9]WorkoutRound
}

type ExerciseTimes struct {
	ExercisePerSet float32
	RestPerSet     float32
	Sets           int
	RestPerRound   float32
	FullRound      float32
	ComboExers     int
}

type WorkoutRound struct {
	ExerciseIDs []string
	Reps        []float32
	Pairs       []bool
	Status      string
	Times       ExerciseTimes
	Rating      float32
}
