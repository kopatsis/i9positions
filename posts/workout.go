package posts

import (
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

	dynamicSets, dynamicNames, dynamicSamples := DynamicSets(dynamics, WOBody.Dynamics, WOBody.StretchTimes, resolution, imagesets)
	workout.DynamicSlice = dynamicSets
	workout.DynamicNames = dynamicNames
	workout.DynamicSamples = dynamicSamples

	staticSets, staticNames, staticSamples := StaticSets(statics, WOBody.Statics, WOBody.StretchTimes, resolution, imagesets)
	workout.StaticSlice = staticSets
	workout.StaticNames = staticNames
	workout.StaticSamples = staticSamples

	workout.DynamicRest = WOBody.StretchTimes.DynamicRest
	workout.DynamicTime = WOBody.StretchTimes.FullRound
	workout.StaticTime = WOBody.StretchTimes.FullRound

	retExers := [9]datatypes.WORound{}
	for i, round := range WOBody.Exercises {
		currentRound := datatypes.WORound{
			Names:     []string{},
			SampleIDs: []string{},
		}

		currentRound.Type = round.Status
		for _, id := range round.ExerciseIDs {
			currentRound.Names = append(currentRound.Names, exercises[id].Name)
			currentRound.SampleIDs = append(currentRound.Names, exercises[id].SampleID)
		}

		currentRound.SetCount = round.Times.Sets
		currentRound.FullTime = round.Times.FullRound
		currentRound.RestPerRound = round.Times.RestPerRound
		currentRound.RestPerSet = round.Times.RestPerSet

		if round.Status == "Regular" {
			currentRound.SetSlice, currentRound.SetSequence, currentRound.Reps = RegularRound(exercises, round, imagesets, resolution)
		} else if round.Status == "Combo" {
			currentRound.SetSlice, currentRound.SetSequence, currentRound.Reps = ComboRound(exercises, round, imagesets, resolution, matrix)
		} else {
			currentRound.SetSlice, currentRound.SetSequence, currentRound.Reps, currentRound.SplitPairs = SplitRound(exercises, round, imagesets, resolution, matrix)
		}

		currentRound.RestPosition = getSpecific(imagesets, resolution, "resting")

		retExers[i] = currentRound
	}
	workout.Exercises = retExers

	workout.CongratsPosition = getSpecific(imagesets, resolution, "congrats")
	workout.StandingPosition = getSpecific(imagesets, resolution, "standing arms bent")

	workout.BackendID = WOBody.ID.Hex()

	return workout, nil
}

func RegularRound(exercises map[string]datatypes.Exercise, round datatypes.WorkoutRound, imagesets map[string]datatypes.ImageSet, resolution string) ([]datatypes.Set, []int, []int) {

	setSlice, setSequence, roundReps := []datatypes.Set{}, []int{}, []int{}

	exer := exercises[round.ExerciseIDs[0]]

	if len(exer.PositionSlice2) == 0 {
		displayReps := customRound(round.Reps[0])
		if isWhole(displayReps) {

			set := SingleRepSet(exer, displayReps, round.Times.ExercisePerSet, imagesets, resolution)

			setSlice = append(setSlice, set)

			for i := 0; i < round.Times.Sets; i++ {
				setSequence = append(setSequence, 0)
			}

			roundReps = append(roundReps, int(displayReps))

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

			roundReps = append(roundReps, int(repCount1))
			roundReps = append(roundReps, int(repCount2))
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

		roundReps = append(roundReps, int(displayReps))
	}

	return setSlice, setSequence, roundReps
}

func ComboRound(exercises map[string]datatypes.Exercise, round datatypes.WorkoutRound, imagesets map[string]datatypes.ImageSet, resolution string, matrix datatypes.TransitionMatrix) ([]datatypes.Set, []int, []int) {

	setSlice, setSequence, roundReps := []datatypes.Set{}, []int{}, []int{}

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

			roundReps = append(roundReps, int(displayReps))

			remainder = 1 + ((round.Reps[i])-displayReps)/(round.Reps[i])

			setsToCombine = append(setsToCombine, SingleRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution))
		}

		set := combineSets(setsToCombine, transitions)

		set.PositionInit = getInitImageSet(exercises[round.ExerciseIDs[0]], imagesets, resolution)
		set.PositionInit = getInitImageSet(exercises[round.ExerciseIDs[len(round.ExerciseIDs)-1]], imagesets, resolution)

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

			roundReps = append(roundReps, int(displayReps))

			remainder = 1 + ((round.Reps[i])-displayReps)/(round.Reps[i])

			setsToCombine1 = append(setsToCombine1, AlternatingRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution, true))

			if int(displayReps)%2 == 0 {
				setsToCombine2 = append(setsToCombine2, AlternatingRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution, true))
			} else {
				setsToCombine2 = append(setsToCombine2, AlternatingRepSet(exercises[exID], displayReps, perExerTime, imagesets, resolution, false))
			}
		}

		set1, set2 := combineSets(setsToCombine1, transitions), combineSets(setsToCombine2, transitions)

		set1.PositionInit = getInitImageSet(exercises[round.ExerciseIDs[0]], imagesets, resolution)
		set1.PositionInit = getInitImageSet(exercises[round.ExerciseIDs[len(round.ExerciseIDs)-1]], imagesets, resolution)

		set2.PositionInit = getInitImageSet(exercises[round.ExerciseIDs[0]], imagesets, resolution)
		set2.PositionInit = getInitImageSet(exercises[round.ExerciseIDs[len(round.ExerciseIDs)-1]], imagesets, resolution)

		setSlice = []datatypes.Set{set1, set2}

		for i := 0; i < round.Times.Sets; i++ {
			if i%2 == 0 {
				setSequence = append(setSequence, 0)
			} else {
				setSequence = append(setSequence, 1)
			}

		}
	}

	return setSlice, setSequence, roundReps

}

func SplitRound(exercises map[string]datatypes.Exercise, round datatypes.WorkoutRound, imagesets map[string]datatypes.ImageSet, resolution string, matrix datatypes.TransitionMatrix) ([]datatypes.Set, []int, []int, [2]bool) {

	setSlice, setSequence, roundReps := []datatypes.Set{}, []int{}, []int{}
	var pairs [2]bool

	displayReps := customRound(round.Reps[0])
	if isWhole(displayReps) {

		set, pairsRet := splitSet(exercises[round.ExerciseIDs[0]], exercises[round.ExerciseIDs[1]], round.Times.ExercisePerSet, imagesets, resolution, matrix, displayReps)
		pairs = pairsRet

		setSlice = append(setSlice, set)

		for i := 0; i < round.Times.Sets; i++ {
			setSequence = append(setSequence, 0)
		}

		roundReps = append(roundReps, int(displayReps))

	} else {

		repCount1 := float32(math.Floor(float64(displayReps)))
		repCount2 := repCount1 + 1

		set1, _ := splitSet(exercises[round.ExerciseIDs[0]], exercises[round.ExerciseIDs[1]], round.Times.ExercisePerSet, imagesets, resolution, matrix, repCount1)
		set2, pairsRet := splitSet(exercises[round.ExerciseIDs[0]], exercises[round.ExerciseIDs[1]], round.Times.ExercisePerSet, imagesets, resolution, matrix, repCount2)
		pairs = pairsRet

		setSlice = []datatypes.Set{set1, set2}

		for i := 0; i < round.Times.Sets; i++ {
			if i%2 == 0 {
				setSequence = append(setSequence, 0)
			} else {
				setSequence = append(setSequence, 1)
			}

		}

		roundReps = append(roundReps, int(repCount1))
		roundReps = append(roundReps, int(repCount2))
	}

	return setSlice, setSequence, roundReps, pairs
}

func splitSet(exer1, exer2 datatypes.Exercise, exercisePerSet float32, imagesets map[string]datatypes.ImageSet, resolution string, matrix datatypes.TransitionMatrix, displayReps float32) (datatypes.Set, [2]bool) {
	timeGigaRep := exercisePerSet / displayReps

	pairs := [2]bool{false, false}

	trans1, trans2 := getSingleTransition(exer1, exer2, matrix, ""), getSingleTransition(exer2, exer1, matrix, "")

	sumTime := trans1.FullTime + trans2.FullTime

	sumTime += (exer1.MaxSecs + exer1.MinSecs) / 2
	if len(exer1.PositionSlice2) != 0 {
		sumTime += (exer1.MaxSecs + exer1.MinSecs) / 2
	}

	sumTime += (exer2.MaxSecs + exer2.MinSecs) / 2
	if len(exer2.PositionSlice2) != 0 {
		sumTime += (exer2.MaxSecs + exer2.MinSecs) / 2
	}

	transRep1, transRep2 := datatypes.Rep{}, datatypes.Rep{}

	if timeGigaRep >= 1.05*sumTime {
		transRep1 = transitionRepToRep(getSingleTransition(exer1, exer2, matrix, "Slow"), imagesets, resolution)
		transRep2 = transitionRepToRep(getSingleTransition(exer2, exer2, matrix, "Slow"), imagesets, resolution)
	} else if timeGigaRep <= .95*sumTime {
		transRep1 = transitionRepToRep(getSingleTransition(exer1, exer2, matrix, "Fast"), imagesets, resolution)
		transRep2 = transitionRepToRep(getSingleTransition(exer2, exer2, matrix, "Fast"), imagesets, resolution)
	} else {
		transRep1 = transitionRepToRep(trans1, imagesets, resolution)
		transRep2 = transitionRepToRep(trans2, imagesets, resolution)
	}

	justExerTime := timeGigaRep - transRep1.FullTime - transRep2.FullTime
	sumTime -= (trans1.FullTime + trans2.FullTime)

	exer1defaultTime := (exer1.MaxSecs + exer1.MinSecs) / 2
	if len(exer1.PositionSlice2) != 0 {
		exer1defaultTime += (exer1.MaxSecs + exer1.MinSecs) / 2
	}

	exer2defaultTime := (exer2.MaxSecs + exer2.MinSecs) / 2
	if len(exer2.PositionSlice2) != 0 {
		exer2defaultTime += (exer2.MaxSecs + exer2.MinSecs) / 2
	}

	exer1RealTime := (exer1defaultTime / sumTime) * justExerTime
	exer2RealTime := (exer2defaultTime / sumTime) * justExerTime

	exer1Rep := exerToRep(exer1, exer1RealTime, imagesets, resolution, false)
	if len(exer1.PositionSlice2) != 0 {
		exer1Rep = combineReps(exer1Rep, exerToRep(exer1, exer1RealTime, imagesets, resolution, true))
		pairs[0] = true
	}

	exer2Rep := exerToRep(exer2, exer2RealTime, imagesets, resolution, false)
	if len(exer2.PositionSlice2) != 0 {
		exer2Rep = combineReps(exer2Rep, exerToRep(exer2, exer2RealTime, imagesets, resolution, true))
		pairs[1] = true
	}

	exer1WTrans := combineReps(exer1Rep, transRep1)
	exer2WTrans := combineReps(exer2Rep, transRep2)

	gigaRep := combineReps(exer1WTrans, exer2WTrans)

	set := datatypes.Set{
		RepSlice:    []datatypes.Rep{gigaRep},
		RepCount:    int(displayReps),
		RepSequence: []int{},
		FullTime:    0,
	}

	for i := 0; i < set.RepCount; i++ {
		set.RepSequence = append(set.RepSequence, 0)
		set.FullTime += gigaRep.FullTime
	}

	set.PositionInit = getInitImageSet(exer1, imagesets, resolution)
	set.PositionEnd = getInitImageSet(exer2, imagesets, resolution)

	return set, pairs

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

	set.PositionInit = getInitImageSet(exer, imagesets, resolution)
	set.PositionEnd = getInitImageSet(exer, imagesets, resolution)

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

	set.PositionInit = getInitImageSet(exer, imagesets, resolution)
	set.PositionEnd = getInitImageSet(exer, imagesets, resolution)

	return set
}

func combineSets(sets []datatypes.Set, transitions []datatypes.Rep) datatypes.Set {
	ret := sets[0]

	for i, set := range sets {
		if i != 0 {
			transition := transitions[i-1]

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

func combineReps(rep1 datatypes.Rep, rep2 datatypes.Rep) datatypes.Rep {
	return datatypes.Rep{
		Positions: append(rep1.Positions, rep2.Positions...),
		Times:     append(rep1.Times, rep2.Times...),
		FullTime:  rep1.FullTime + rep2.FullTime,
	}
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

func exerToRep(exer datatypes.Exercise, initRepTime float32, imagesets map[string]datatypes.ImageSet, resolution string, alternate bool) datatypes.Rep {
	rep := datatypes.Rep{}
	positions, times := [][]string{}, []float32{}
	calcTime := float32(math.Max(float64(initRepTime), float64(exer.MinSecs)))

	if !alternate {
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
	} else {
		for _, position := range exer.PositionSlice2 {
			if position.Hardcoded {
				calcTime -= position.HardcodedSecs
			}
		}

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

	rep.Positions = positions
	rep.Times = times
	rep.FullTime = initRepTime

	return rep
}

func getSingleTransition(exer1, exer2 datatypes.Exercise, matrix datatypes.TransitionMatrix, speed string) datatypes.TransitionRep {
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

	switch speed {
	case "Slow":
		return matrix.SlowMatrix[parentMatIndex[exer1.Parent]][parentMatIndex[exer2.Parent]]
	case "Fast":
		return matrix.FastMatrix[parentMatIndex[exer1.Parent]][parentMatIndex[exer2.Parent]]
	default:
		return matrix.RegularMatrix[parentMatIndex[exer1.Parent]][parentMatIndex[exer2.Parent]]
	}
}

func transitionRepToRep(transition datatypes.TransitionRep, imagesets map[string]datatypes.ImageSet, resolution string) datatypes.Rep {
	rep := datatypes.Rep{
		FullTime:  transition.FullTime,
		Times:     transition.Times,
		Positions: [][]string{},
	}

	switch resolution {
	case "Low":
		for _, id := range transition.ImageSetIDs {
			rep.Positions = append(rep.Positions, imagesets[id].Low)
		}
	case "Mid":
		for _, id := range transition.ImageSetIDs {
			rep.Positions = append(rep.Positions, imagesets[id].Mid)
		}
	case "High":
		for _, id := range transition.ImageSetIDs {
			rep.Positions = append(rep.Positions, imagesets[id].High)
		}
	default:
		for _, id := range transition.ImageSetIDs {
			rep.Positions = append(rep.Positions, imagesets[id].Original)
		}
	}

	return rep
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

func getInitImageSet(exer datatypes.Exercise, imagesets map[string]datatypes.ImageSet, resolution string) []string {
	switch resolution {
	case "Low":
		return imagesets[exer.ImageSetID0].Low
	case "Mid":
		return imagesets[exer.ImageSetID0].Mid
	case "High":
		return imagesets[exer.ImageSetID0].High
	default:
		return imagesets[exer.ImageSetID0].Original
	}
}

func isWhole(value float32) bool {
	const tolerance = 0.001
	diff := float64(value) - math.Round(float64(value))
	return math.Abs(diff) < tolerance
}
