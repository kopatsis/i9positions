package posts

import (
	"errors"
	"i9-pos/database"
	"i9-pos/datatypes"
	"math"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

func StretchWorkout(db *mongo.Database, resolution string, strWOBody datatypes.StretchWorkoutRoute) (datatypes.StretchWorkout, error) {

	retWO := datatypes.StretchWorkout{}

	dynamics, statics, imagesets, err := database.QueryStretchWO(db, strWOBody.Statics, strWOBody.Dynamics)
	if err != nil {
		return datatypes.StretchWorkout{}, err
	}

	if len(dynamics) == 0 || len(statics) == 0 || len(imagesets) == 0 {
		return datatypes.StretchWorkout{}, errors.New("unfilled dynamic/static/imagesets returned")
	}

	dynamicSets, dynamicNames, dynamicSamples := DynamicSets(dynamics, strWOBody.Dynamics, strWOBody.StretchTimes, resolution, imagesets)
	retWO.DynamicSlice = dynamicSets
	retWO.DynamicNames = dynamicNames
	retWO.DynamicSamples = dynamicSamples

	staticSets, staticNames, staticSamples := StaticSets(statics, strWOBody.Statics, strWOBody.StretchTimes, resolution, imagesets)
	retWO.StaticSlice = staticSets
	retWO.StaticNames = staticNames
	retWO.StaticSamples = staticSamples

	retWO.RoundTime = strWOBody.StretchTimes.FullRound / 2

	retWO.CongratsPosition = getSpecific(imagesets, resolution, "congrat")
	retWO.StandingPosition = getSpecific(imagesets, resolution, "standing arms bent")

	retWO.BackendID = strWOBody.ID.Hex()

	return retWO, nil
}

func StaticSets(statics map[string]datatypes.StaticStr, staticList []string, stretchTimes datatypes.StretchTimes, resolution string, imagesets map[string]datatypes.ImageSet) ([]datatypes.Set, []string, []string) {
	staticSets := []datatypes.Set{}
	staticNames := []string{}
	staticSamples := []string{}

	for i, id := range staticList {
		static, set := statics[id], datatypes.Set{}

		if static.ImageSetID2 == "" {
			rep := datatypes.Rep{
				FullTime: stretchTimes.StaticPerSet[i],
				Times:    []float32{stretchTimes.StaticPerSet[i]},
			}

			switch resolution {
			case "Low":
				rep.Positions = [][]string{imagesets[static.ImageSetID1].Low}
			case "Mid":
				rep.Positions = [][]string{imagesets[static.ImageSetID1].Mid}
			case "High":
				rep.Positions = [][]string{imagesets[static.ImageSetID1].High}
			default:
				rep.Positions = [][]string{imagesets[static.ImageSetID1].Original}
			}

			set.FullTime = stretchTimes.StaticPerSet[i]
			set.RepCount = 1
			set.RepSlice = []datatypes.Rep{rep}
			set.RepSequence = []int{0}
		} else {
			rep1, rep2 := datatypes.Rep{}, datatypes.Rep{}

			rep1.Times, rep2.Times = []float32{stretchTimes.StaticPerSet[i] / 2}, []float32{stretchTimes.StaticPerSet[i] / 2}
			rep1.FullTime, rep2.FullTime = stretchTimes.StaticPerSet[i]/2, stretchTimes.StaticPerSet[i]/2

			switch resolution {
			case "Low":
				rep1.Positions = [][]string{imagesets[static.ImageSetID1].Low}
				rep2.Positions = [][]string{imagesets[static.ImageSetID2].Low}
			case "Mid":
				rep1.Positions = [][]string{imagesets[static.ImageSetID1].Mid}
				rep2.Positions = [][]string{imagesets[static.ImageSetID2].Mid}
			case "High":
				rep1.Positions = [][]string{imagesets[static.ImageSetID1].High}
				rep2.Positions = [][]string{imagesets[static.ImageSetID2].High}
			default:
				rep1.Positions = [][]string{imagesets[static.ImageSetID1].Original}
				rep2.Positions = [][]string{imagesets[static.ImageSetID2].Original}
			}

			set.FullTime = stretchTimes.StaticPerSet[i]
			set.RepCount = 2
			set.RepSlice = []datatypes.Rep{rep1, rep2}
			set.RepSequence = []int{0, 1}
		}

		staticSets = append(staticSets, set)
		staticNames = append(staticNames, static.Name)
		staticSamples = append(staticSamples, static.SampleID)

	}

	return staticSets, staticNames, staticSamples
}

func DynamicSets(dynamics map[string]datatypes.DynamicStr, dynamicList []string, stretchTimes datatypes.StretchTimes, resolution string, imagesets map[string]datatypes.ImageSet) ([]datatypes.Set, []string, []string) {
	dynamicSets := []datatypes.Set{}
	dynamicNames := []string{}
	dynamicSamples := []string{}

	for i, id := range dynamicList {

		dynamic, set := dynamics[id], datatypes.Set{}

		if len(dynamic.PositionSlice2) == 0 {
			setTime := stretchTimes.DynamicPerSet[i]
			repCount := int(math.Round(float64(setTime) / float64(dynamic.Secs)))
			realRepTime := setTime / float32(repCount)

			var currentRep datatypes.Rep
			currentRep.FullTime = realRepTime
			positions, times := [][]string{}, []float32{}

			for _, position := range dynamic.PositionSlice1 {
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
				times = append(times, position.PercentSecs*realRepTime)
			}

			currentRep.Positions = positions
			currentRep.Times = times

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
				repCount = int(math.RoundToEven(float64(setTime) / float64(dynamic.Secs)))
			} else {
				repCount = int(math.Round(float64(setTime) / float64(dynamic.Secs)))
			}

			realRepTime := setTime / float32(repCount)

			rep1, rep2 := datatypes.Rep{}, datatypes.Rep{}
			rep1.FullTime = realRepTime
			rep2.FullTime = realRepTime

			positions, times := [][]string{}, []float32{}

			for _, position := range dynamic.PositionSlice1 {
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
				times = append(times, position.PercentSecs*realRepTime)
			}

			rep1.Positions = positions
			rep1.Times = times

			positions, times = [][]string{}, []float32{}

			for _, position := range dynamic.PositionSlice2 {
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

func getSpecific(imagesets map[string]datatypes.ImageSet, resolution, includes string) []string {
	for _, imageset := range imagesets {
		if strings.Contains(strings.ToLower(imageset.Name), strings.ToLower(includes)) {
			switch resolution {
			case "Low":
				return imageset.Low
			case "Mid":
				return imageset.Mid
			case "High":
				return imageset.High
			default:
				return imageset.Original
			}
		}
	}
	return []string{}
}
