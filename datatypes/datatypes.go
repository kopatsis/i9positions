package datatypes

import "go.mongodb.org/mongo-driver/bson/primitive"

// Exists in DB as actual entry
type ImageSet struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Low      []string           `bson:"low"`
	Mid      []string           `bson:"mid"`
	High     []string           `bson:"high"`
	Original []string           `bson:"original"`
	Name     string             `bson:"name"`
}

// Exists in DB as part of other entry
type ExerPosition struct {
	ImageSetID    string  `bson:"imageset"`
	Hardcoded     bool    `bson:"hardcoded"`
	HardcodedSecs float32 `bson:"hardcodedsecs"`
	MaxSecs       float32 `bson:"maxsecs"` // ?
	PercentSecs   float32 `bson:"percentsecs"`
}

// Exists in DB as part of other entry
type StrPosition struct {
	ImageSetID  string  `bson:"imageset"`
	PercentSecs float32 `bson:"percentsecs"`
}

// Exists in DB as actual entry
type Exercise struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	BackendID      string             `bson:"backendID"`
	Name           string             `bson:"name"`
	MaxSecs        float32            `bson:"maxsecs"`
	MinSecs        float32            `bson:"minsecs"`
	PositionSlice1 []ExerPosition     `bson:"positions1"`
	PositionSlice2 []ExerPosition     `bson:"positions2"`
	SampleID       string             `bson:"sampleid"`
}

// Exists in DB as actual entry
type DynamicStr struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	BackendID      string             `bson:"backendID"`
	Name           string             `bson:"name"`
	Secs           float32            `bson:"secs"`
	PositionSlice1 []StrPosition      `bson:"positions1"`
	PositionSlice2 []StrPosition      `bson:"positions2"`
	SampleID       string             `bson:"sampleid"`
}

// Exists in DB as actual entry
type StaticStr struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	BackendID   string             `bson:"backendID"`
	Name        string             `bson:"name"`
	ImageSetID1 string             `bson:"imageset1"`
	ImageSetID2 string             `bson:"imageset2"`
	SampleID    string             `bson:"sampleid"`
}

// Exists in DB as part of other entry
type Rep struct {
	Positions [][]string `bson:"positions"`
	Times     []float32  `bson:"times"`
	FullTime  float32    `bson:"fulltime"`
}

// Exists in DB as part of other entry
type Set struct {
	RepSlice    []Rep   `bson:"reps"`
	RepSequence []int   `bson:"repsequence"`
	RepCount    int     `bson:"repcount"`
	FullTime    float32 `bson:"fulltime"`
}

// Exists in DB as part of other entry
type WORound struct {
	SetSlice     []Set   `bson:"sets"`
	SetSequence  []int   `bson:"setsequence"`
	SetCount     int     `bson:"setcount"`
	FullTime     float32 `bson:"fulltime"`
	RestPerRound float32 `bson:"restround"`
	RestPerSet   float32 `bson:"restset"`
}

// Exists in DB as actual entry
type Sample struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	RepSlice    []Rep              `bson:"reps"`
}

// Exists in DB as actual entry
type StretchWorkout struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	BackendID    string             `bson:"backendID"`
	DynamicSlice []Set              `bson:"dynamics"`
	StaticSlice  []Set              `bson:"statics"`
	RoundTime    float32            `bson:"roundtime"`
}

// Exists in DB as actual entry
type Workout struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	BackendID    string             `bson:"backendID"`
	DynamicSlice []Set              `bson:"dynamics"`
	StaticSlice  []Set              `bson:"statics"`
	StaticTime   float32            `bson:"statictime"`
	DynamicTime  float32            `bson:"dynamictime"`
	DynamicRest  float32            `bson:"dynamicrest"`
	Exercises    [9]WORound         `bson:"exercises"`
}
