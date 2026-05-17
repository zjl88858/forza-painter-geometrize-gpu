package model

type Settings struct {
	Description     string
	MaxPreviewSize  int
	MaxResolution   int
	MaxThreads      int
	MutatedSamples  int
	PosterizeLevels int
	PreviewEvery    int
	RandomSamples   int
	SaveAt          map[int]struct{}
	SaveEvery       int
	StopAt          int
}

type Shape struct {
	Type  int     `json:"type"`
	Data  []int   `json:"data"`
	Color []int   `json:"color"`
	Score float64 `json:"score"`
}

type Geometry struct {
	Shapes []Shape `json:"shapes"`
}

type Candidate struct {
	X     float32
	Y     float32
	RX    float32
	RY    float32
	Theta float32
	R     float32
	G     float32
	B     float32
	A     float32
}
