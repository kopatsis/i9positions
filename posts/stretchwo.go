package posts

import (
	"errors"
	"i9-pos/database"
	"i9-pos/datatypes"
	"math"

	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/mongo"
)

func StretchWorkout(db *mongo.Database, boltDB *bbolt.DB, strWOBody datatypes.StretchWorkoutRoute) (datatypes.StretchWorkout, error) {

	retWO := datatypes.StretchWorkout{}

	dynamics, statics, err := database.QueryStretchWO(db, boltDB, strWOBody.Statics, strWOBody.Dynamics)
	if err != nil {
		return datatypes.StretchWorkout{}, err
	}

	if len(dynamics) == 0 || len(statics) == 0 {
		return datatypes.StretchWorkout{}, errors.New("unfilled dynamic/static/imagesets returned")
	}

	dynamicSets, dynamicNames, dynamicSamples := DynamicSets(dynamics, strWOBody.Dynamics, strWOBody.StretchTimes)
	retWO.DynamicSlice = dynamicSets
	retWO.DynamicNames = dynamicNames
	retWO.DynamicSamples = dynamicSamples

	staticSets, staticNames, staticSamples := StaticSets(statics, strWOBody.Statics, strWOBody.StretchTimes)
	retWO.StaticSlice = staticSets
	retWO.StaticNames = staticNames
	retWO.StaticSamples = staticSamples

	retWO.RoundTime = strWOBody.StretchTimes.FullRound / 2

	retWO.CongratsPosition = "standing-thumbs-up-wink"
	retWO.StandingPosition = "standing-arms-bent"

	retWO.BackendID = strWOBody.ID.Hex()

	return retWO, nil
}

func StaticSets(statics map[string]datatypes.StaticStr, staticList []string, stretchTimes datatypes.StretchTimes) ([]datatypes.Set, []string, []string) {
	staticSets := []datatypes.Set{}
	staticNames := []string{}
	staticSamples := []string{}

	for i, id := range staticList {
		static, set := statics[id], datatypes.Set{}

		if static.ImageSetID2 == "" {
			rep := datatypes.Rep{
				FullTime:  stretchTimes.StaticPerSet[i],
				Times:     []float32{stretchTimes.StaticPerSet[i]},
				Positions: []string{static.ImageSetID1},
			}

			set.FullTime = stretchTimes.StaticPerSet[i]
			set.RepCount = 1
			set.RepSlice = []datatypes.Rep{rep}
			set.RepSequence = []int{0}
			set.SeparateStretch = false
		} else {
			rep1, rep2 := datatypes.Rep{}, datatypes.Rep{}

			rep1.Times, rep2.Times = []float32{stretchTimes.StaticPerSet[i] / 2}, []float32{stretchTimes.StaticPerSet[i] / 2}
			rep1.FullTime, rep2.FullTime = stretchTimes.StaticPerSet[i]/2, stretchTimes.StaticPerSet[i]/2
			rep1.Positions, rep2.Positions = []string{static.ImageSetID1}, []string{static.ImageSetID2}

			set.FullTime = stretchTimes.StaticPerSet[i]
			set.RepCount = 2
			set.RepSlice = []datatypes.Rep{rep1, rep2}
			set.RepSequence = []int{0, 1}
			set.SeparateStretch = true
		}

		staticSets = append(staticSets, set)
		staticNames = append(staticNames, static.Name)
		staticSamples = append(staticSamples, static.SampleID)

	}

	return staticSets, staticNames, staticSamples
}

func DynamicSets(dynamics map[string]datatypes.DynamicStr, dynamicList []string, stretchTimes datatypes.StretchTimes) ([]datatypes.Set, []string, []string) {
	dynamicSets := []datatypes.Set{}
	dynamicNames := []string{}
	dynamicSamples := []string{}

	for i, id := range dynamicList {

		dynamic, set := dynamics[id], datatypes.Set{}

		if len(dynamic.PositionSlice2) == 0 {
			setTime := stretchTimes.DynamicPerSet[i]
			repCount := int(math.Max(1, math.Round(float64(setTime)/float64(dynamic.Secs))))
			realRepTime := setTime / float32(repCount)

			var currentRep datatypes.Rep
			currentRep.FullTime = realRepTime
			positions, times := []string{}, []float32{}

			for _, position := range dynamic.PositionSlice1 {
				positions = append(positions, position.ImageSetID)
				times = append(times, position.PercentSecs*realRepTime)
			}

			currentRep.Positions = positions
			currentRep.Times = times

			set.SeparateStretch = false
			set.FullTime = setTime
			set.RepSlice = []datatypes.Rep{currentRep}
			set.RepCount = repCount
			set.RepSequence = []int{}
			for i := 0; i < repCount; i++ {
				set.RepSequence = append(set.RepSequence, 0)
			}

		} else {

			setTime := stretchTimes.DynamicPerSet[i]

			var repCount int
			if dynamic.SeparateSets {
				repCount = int(math.Max(1, math.RoundToEven(float64(setTime)/float64(dynamic.Secs))))
				set.SeparateStretch = true
			} else {
				repCount = int(math.Max(1, math.Round(float64(setTime)/float64(dynamic.Secs))))
				set.SeparateStretch = false
			}

			realRepTime := setTime / float32(repCount)

			rep1, rep2 := datatypes.Rep{}, datatypes.Rep{}
			rep1.FullTime = realRepTime
			rep2.FullTime = realRepTime

			positions, times := []string{}, []float32{}

			for _, position := range dynamic.PositionSlice1 {
				positions = append(positions, position.ImageSetID)
				times = append(times, position.PercentSecs*realRepTime)
			}

			rep1.Positions = positions
			rep1.Times = times

			positions, times = []string{}, []float32{}

			for _, position := range dynamic.PositionSlice2 {
				positions = append(positions, position.ImageSetID)
				times = append(times, position.PercentSecs*realRepTime)
			}

			rep2.Positions = positions
			rep2.Times = times

			set.FullTime = setTime
			set.RepSlice = []datatypes.Rep{rep1, rep2}
			set.RepCount = repCount
			set.RepSequence = []int{}

			if !dynamic.SeparateSets {
				for i := 0; i < repCount; i++ {
					if i%2 == 0 {
						set.RepSequence = append(set.RepSequence, 0)
					} else {
						set.RepSequence = append(set.RepSequence, 1)
					}

				}
			} else {
				for i := 0; i < repCount; i++ {
					if i*2 < repCount {
						set.RepSequence = append(set.RepSequence, 0)
					} else {
						set.RepSequence = append(set.RepSequence, 1)
					}

				}
			}

		}

		dynamicSets = append(dynamicSets, set)
		dynamicNames = append(dynamicNames, dynamic.Name)
		dynamicSamples = append(dynamicSamples, dynamic.SampleID)

	}

	return dynamicSets, dynamicNames, dynamicSamples
}
