package posts

import (
	"fmt"
	"i9-pos/database"
	"i9-pos/datatypes"
	"math"

	"go.mongodb.org/mongo-driver/mongo"
)

func Workout(db *mongo.Database, resolution string, WOBody datatypes.WorkoutRoute) (datatypes.Workout, error) {
	workout := datatypes.Workout{}

	exerIDRoundList := [9][]string{}
	for i, workoutRound := range WOBody.Exercises {
		exerIDRoundList[i] = workoutRound.ExerciseIDs
	}

	dynamics, statics, imagesets, exercises, matrix, err := database.QueryWO(db, WOBody.Statics, WOBody.Dynamics, exerIDRoundList)
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
			currentRound.SetSlice, currentRound.SetSequence = RegularRound(exercises, round, imagesets, resolution)
		} else if round.Status == "Combo" {
			fmt.Println(matrix)
		} else {
			fmt.Println(matrix)
		}

		retExers[i] = currentRound
	}
	workout.Exercises = retExers

	workout.BackendID = WOBody.ID.Hex()

	return workout, nil
}

func RegularRound(exercises map[string]datatypes.Exercise, round datatypes.WorkoutRound, imagesets map[string]datatypes.ImageSet, resolution string) ([]datatypes.Set, []int) {

	setSlice, setSequence := []datatypes.Set{}, []int{}

	exer := exercises[round.ExerciseIDs[0]]

	if len(exer.PositionSlice2) == 0 {
		displayReps := customRound(round.Reps[0])
		if !(math.Mod(float64(displayReps), 1) > 0.4) {

			set := SingleRepSet(exer, displayReps, round.Times.ExercisePerSet, imagesets, resolution)

			setSlice = append(setSlice, set)

			for i := 0; i < round.Times.Sets; i++ {
				setSequence = append(setSequence, 0)
			}

		} else {

			repCount1 := float32(math.Floor(float64(displayReps)))
			repCount2 := repCount1 + 1

			set1 := SingleRepSet(exer, repCount1, round.Times.ExercisePerSet, imagesets, resolution)
			set2 := SingleRepSet(exer, repCount2, round.Times.ExercisePerSet, imagesets, resolution)

			setSlice = []datatypes.Set{set1, set2}

			for i := 0; i < round.Times.Sets; i++ {
				if i%2 == 0 {
					setSequence = append(setSequence, 0)
				} else {
					setSequence = append(setSequence, 1)
				}

			}
		}
	} else {
		displayReps := float32(math.Round(float64(round.Reps[0])))

		if int(displayReps)%2 == 0 {

			set := AlternatingRepSet(exer, displayReps, round.Times.ExercisePerSet, imagesets, resolution, false)

			setSlice = append(setSlice, set)

			for i := 0; i < round.Times.Sets; i++ {
				setSequence = append(setSequence, 0)
			}

		} else {

			set1 := AlternatingRepSet(exer, displayReps, round.Times.ExercisePerSet, imagesets, resolution, false)
			set2 := AlternatingRepSet(exer, displayReps, round.Times.ExercisePerSet, imagesets, resolution, true)

			setSlice = []datatypes.Set{set1, set2}

			for i := 0; i < round.Times.Sets; i++ {
				if i%2 == 0 {
					setSequence = append(setSequence, 0)
				} else {
					setSequence = append(setSequence, 1)
				}

			}

		}
	}

	return setSlice, setSequence
}

func ComboRound(exercises map[string]datatypes.Exercise, round datatypes.WorkoutRound, imagesets map[string]datatypes.ImageSet, resolution string, matrix datatypes.TransitionMatrix) ([]datatypes.Set, []int) {

	setSlice, setSequence := []datatypes.Set{}, []int{}

	hasDoubles := false

	for _, exID := range round.ExerciseIDs {
		if len(exercises[exID].PositionSlice2) != 0 {
			hasDoubles = true
		}
	}

	if !hasDoubles {
		transitions, workingTime := getTransitions(exercises, round, imagesets, resolution, matrix)

		perExerTime := workingTime / float32(round.Times.ComboExers)

		setsToCombine := []datatypes.Set{}
		remainder := float32(1.0)

		for i, exID := range round.ExerciseIDs {
			displayReps := float32(math.Round(float64(perExerTime/round.Reps[i]))) * remainder

			remainder = 1 + ((round.Reps[i])-displayReps)/(round.Reps[i])

			setsToCombine = append(setsToCombine, SingleRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution))
		}

		set := combineSets(setsToCombine, transitions)

		setSlice = append(setSlice, set)

		for i := 0; i < round.Times.Sets; i++ {
			setSequence = append(setSequence, 0)
		}

	} else {
		transitions, workingTime := getTransitions(exercises, round, imagesets, resolution, matrix)

		perExerTime := workingTime / float32(round.Times.ComboExers)

		setsToCombine1, setsToCombine2 := []datatypes.Set{}, []datatypes.Set{}
		remainder := float32(1.0)

		for i, exID := range round.ExerciseIDs {
			displayReps := float32(math.Round(float64(perExerTime/round.Reps[i]))) * remainder

			remainder = 1 + ((round.Reps[i])-displayReps)/(round.Reps[i])

			setsToCombine1 = append(setsToCombine1, AlternatingRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution, true))

			if int(displayReps)%2 == 0 {
				setsToCombine2 = append(setsToCombine2, AlternatingRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution, true))
			} else {
				setsToCombine2 = append(setsToCombine2, AlternatingRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution, false))
			}
		}

		set1, set2 := combineSets(setsToCombine1, transitions), combineSets(setsToCombine2, transitions)

		setSlice = []datatypes.Set{set1, set2}

		for i := 0; i < round.Times.Sets; i++ {
			if i%2 == 0 {
				setSequence = append(setSequence, 0)
			} else {
				setSequence = append(setSequence, 1)
			}

		}
	}

	return setSlice, setSequence

}

func SingleRepSet(exer datatypes.Exercise, displayReps float32, exercisePerSet float32, imagesets map[string]datatypes.ImageSet, resolution string) datatypes.Set {

	rep, set := datatypes.Rep{}, datatypes.Set{}
	positions, times := [][]string{}, []float32{}
	initRepTime := float32(math.Min(float64(exer.MaxSecs), float64(exercisePerSet/displayReps)))
	totalTime := initRepTime * displayReps
	calcTime := float32(math.Max(float64(initRepTime), float64(exer.MinSecs)))

	for _, position := range exer.PositionSlice1 {
		if position.Hardcoded {
			calcTime -= position.HardcodedSecs
		}
	}

	for _, position := range exer.PositionSlice1 {
		if position.Hardcoded {
			times = append(times, position.HardcodedSecs)
		} else {
			times = append(times, calcTime*position.PercentSecs)
		}
		switch resolution {
		case "Low":
			positions = append(positions, imagesets[position.ImageSetID].Low)
		case "Mid":
			positions = append(positions, imagesets[position.ImageSetID].Mid)
		case "High":
			positions = append(positions, imagesets[position.ImageSetID].High)
		default:
			positions = append(positions, imagesets[position.ImageSetID].Original)
		}
	}

	if exer.MinSecs > initRepTime {
		for i, time := range times {
			times[i] = time * (initRepTime / exer.MinSecs)
		}
	}

	rep.Positions = positions
	rep.Times = times
	rep.FullTime = initRepTime

	set.RepSlice = []datatypes.Rep{rep}

	set.RepCount = int(displayReps)

	set.RepSequence = []int{}
	for i := 0; i < set.RepCount; i++ {
		set.RepSequence = append(set.RepSequence, 0)
	}

	set.FullTime = totalTime

	return set
}

func AlternatingRepSet(exer datatypes.Exercise, displayReps float32, exercisePerSet float32, imagesets map[string]datatypes.ImageSet, resolution string, orderReverse bool) datatypes.Set {

	rep1, rep2 := datatypes.Rep{}, datatypes.Rep{}

	initRepTime := float32(math.Min(float64(exer.MaxSecs), float64(exercisePerSet/displayReps)))
	for i := 0; i < 2; i++ {

		positions, times := [][]string{}, []float32{}

		calcTime := float32(math.Max(float64(initRepTime), float64(exer.MinSecs)))

		for _, position := range exer.PositionSlice1 {
			if position.Hardcoded {
				calcTime -= position.HardcodedSecs
			}
		}

		if i == 0 {
			for _, position := range exer.PositionSlice1 {
				if position.Hardcoded {
					times = append(times, position.HardcodedSecs)
				} else {
					times = append(times, calcTime*position.PercentSecs)
				}
				switch resolution {
				case "Low":
					positions = append(positions, imagesets[position.ImageSetID].Low)
				case "Mid":
					positions = append(positions, imagesets[position.ImageSetID].Mid)
				case "High":
					positions = append(positions, imagesets[position.ImageSetID].High)
				default:
					positions = append(positions, imagesets[position.ImageSetID].Original)
				}
			}
		} else {
			for _, position := range exer.PositionSlice2 {
				if position.Hardcoded {
					times = append(times, position.HardcodedSecs)
				} else {
					times = append(times, calcTime*position.PercentSecs)
				}
				switch resolution {
				case "Low":
					positions = append(positions, imagesets[position.ImageSetID].Low)
				case "Mid":
					positions = append(positions, imagesets[position.ImageSetID].Mid)
				case "High":
					positions = append(positions, imagesets[position.ImageSetID].High)
				default:
					positions = append(positions, imagesets[position.ImageSetID].Original)
				}
			}
		}

		if exer.MinSecs > initRepTime {
			for i, time := range times {
				times[i] = time * (initRepTime / exer.MinSecs)
			}
		}

		if i == 0 {
			rep1.Positions = positions
			rep1.Times = times
			rep1.FullTime = initRepTime
		} else {
			rep2.Positions = positions
			rep2.Times = times
			rep2.FullTime = initRepTime
		}

	}

	totalTime := initRepTime * displayReps

	set := datatypes.Set{}
	if orderReverse {
		set.RepSlice = []datatypes.Rep{rep1, rep2}
	} else {
		set.RepSlice = []datatypes.Rep{rep2, rep1}

	}

	set.RepCount = int(displayReps)

	set.RepSequence = []int{}
	for i := 0; i < set.RepCount; i++ {
		if i%2 == 0 {
			set.RepSequence = append(set.RepSequence, 0)
		} else {
			set.RepSequence = append(set.RepSequence, 1)
		}

	}
	set.FullTime = totalTime

	return set
}

func combineSets(sets []datatypes.Set, transitions []datatypes.Rep) datatypes.Set {
	ret := sets[0]

	for i, set := range sets {
		if i != 0 {
			transition := transitions[i]

			originalCt := len(ret.RepSlice)

			ret.RepSlice = append(ret.RepSlice, transition)
			ret.RepSlice = append(ret.RepSlice, set.RepSlice...)

			ret.RepSequence = append(ret.RepSequence, originalCt)

			for _, seq := range set.RepSequence {
				ret.RepSequence = append(ret.RepSequence, seq+originalCt+1)
			}

			ret.RepCount += 1 + set.RepCount
			ret.FullTime += transition.FullTime + set.FullTime
		}
	}

	return ret
}

func customRound(num float32) float32 {
	whole, decimal := math.Modf(float64(num))
	if decimal < 0.35 {
		return float32(whole)
	} else if decimal > 0.65 {
		return float32(whole + 1)
	} else {
		return float32(whole + 0.5)
	}
}

func getTransitions(exercises map[string]datatypes.Exercise, round datatypes.WorkoutRound, imagesets map[string]datatypes.ImageSet, resolution string, matrix datatypes.TransitionMatrix) ([]datatypes.Rep, float32) {

	parentMatIndex := map[string]int{
		"Pushups":           0,
		"Squats":            1,
		"Burpees":           2,
		"Jumps":             3,
		"Lunges":            4,
		"Mountain Climbers": 5,
		"Abs":               6,
		"Bridges":           7,
		"Kicks":             8,
		"Planks":            9,
		"Supermans":         10,
	}

	transitions := []datatypes.Rep{}
	workingTime := round.Times.ExercisePerSet

	for i, exID := range round.ExerciseIDs {
		if i != 0 {
			index1, index2 := parentMatIndex[exercises[round.ExerciseIDs[i-1]].Parent], parentMatIndex[exercises[exID].Parent]
			transRep := matrix.RegularMatrix[index1][index2]

			rep := datatypes.Rep{
				Positions: [][]string{},
				Times:     transRep.Times,
				FullTime:  transRep.FullTime,
			}

			for _, img := range transRep.ImageSetIDs {
				switch resolution {
				case "Low":
					rep.Positions = append(rep.Positions, imagesets[img].Low)
				case "Mid":
					rep.Positions = append(rep.Positions, imagesets[img].Mid)
				case "High":
					rep.Positions = append(rep.Positions, imagesets[img].High)
				default:
					rep.Positions = append(rep.Positions, imagesets[img].Original)
				}
			}

			transitions = append(transitions, rep)
			workingTime -= matrix.RegularMatrix[index1][index2].FullTime
		}
	}

	return transitions, workingTime
}
