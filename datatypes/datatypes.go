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
	Parent         string             `bson:"parent"`
	MaxSecs        float32            `bson:"maxsecs"`
	MinSecs        float32            `bson:"minsecs"`
	ImageSetID0    string             `bson:"imageset0"`
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
	SeparateSets   bool               `bson:"separate"`
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

// Programatically created as actual entry in DB
type Set struct {
	RepSlice     []Rep    `bson:"reps"`
	RepSequence  []int    `bson:"repsequence"`
	RepCount     int      `bson:"repcount"`
	PositionInit []string `bson:"positioninit"`
	PositionEnd  []string `bson:"positionend"`
	FullTime     float32  `bson:"fulltime"`
}

// Programatically created as actual entry in DB
type WORound struct {
	SetSlice     []Set    `bson:"sets"`
	SetSequence  []int    `bson:"setsequence"`
	SetCount     int      `bson:"setcount"`
	FullTime     float32  `bson:"fulltime"`
	RestPerRound float32  `bson:"restround"`
	RestPerSet   float32  `bson:"restset"`
	Type         string   `bson:"type"`
	Names        []string `bson:"names"`
	Reps         []int    `bson:"reps"`
	SplitPairs   [2]bool  `bson:"splitpairs"`
	SampleIDs    []string `bson:"samples"`
	RestPosition []string `bson:"restposition"`
}

// Exists in DB as actual entry
type Sample struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Reps        Rep                `bson:"reps"`
	Type        string             `bson:"type"`
	ExOrStID    string             `bson:"exorstid"`
}

// Programatically created as actual entry in DB
type StretchWorkout struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	BackendID        string             `bson:"backendID"`
	DynamicSlice     []Set              `bson:"dynamics"`
	StaticSlice      []Set              `bson:"statics"`
	DynamicNames     []string           `bson:"dynamicnames"`
	StaticNames      []string           `bson:"staticnames"`
	DynamicSamples   []string           `bson:"dynamicsamples"`
	StaticSamples    []string           `bson:"staticsamples"`
	CongratsPosition []string           `bson:"congratspos"`
	StandingPosition []string           `bson:"standingpos"`
	RoundTime        float32            `bson:"roundtime"`
}

// Programatically created as actual entry in DB
type Workout struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	BackendID        string             `bson:"backendID"`
	DynamicSlice     []Set              `bson:"dynamics"`
	StaticSlice      []Set              `bson:"statics"`
	StaticTime       float32            `bson:"statictime"`
	DynamicTime      float32            `bson:"dynamictime"`
	DynamicRest      float32            `bson:"dynamicrest"`
	DynamicNames     []string           `bson:"dynamicnames"`
	StaticNames      []string           `bson:"staticnames"`
	DynamicSamples   []string           `bson:"dynamicsamples"`
	StaticSamples    []string           `bson:"staticsamples"`
	CongratsPosition []string           `bson:"congratspos"`
	StandingPosition []string           `bson:"standingpos"`
	Exercises        [9]WORound         `bson:"exercises"`
}

type TransitionRep struct {
	ImageSetIDs []string  `bson:"imagesetids"`
	Times       []float32 `bson:"times"`
	FullTime    float32   `bson:"fulltime"`
}

type TransitionMatrix struct {
	ID            primitive.ObjectID    `bson:"_id,omitempty"`
	FastMatrix    [11][11]TransitionRep `bson:"fastmatrix"`
	RegularMatrix [11][11]TransitionRep `bson:"regularmatrix"`
	SlowMatrix    [11][11]TransitionRep `bson:"slowmatrix"`
}
