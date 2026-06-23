package model

type Settings struct {
	Description                   string
	MaxPreviewSize                int
	MaxResolution                 int
	MaxThreads                    int
	MutatedSamples                int
	ForceOpaqueShapes             bool
	EnableMultiPrimitiveShapes    bool
	PosterizeLevels               int
	PreviewEvery                  int
	RandomSamples                 int
	SaveAt                        map[int]struct{}
	SaveEvery                     int
	StopAt                        int
	UseWorkGroupEval              bool
	EnableProgressiveSampling     bool
	ProgressiveSamplingStart      int
	ProgressiveSamplingEnd        int
	ProgressiveSamplingTransition float32
	ProgressiveSamplingCurve      float32
	ErrorGridSize                 int
	LoadGeometry                  string
	ErrorMetric                   string
	SsimWeight                    float32
}

type Shape struct {
	Type  int       `json:"type"`
	Data  []float64 `json:"data"`
	Color []int     `json:"color"`
	Score float64   `json:"score"`
}

type Geometry struct {
	Shapes []Shape `json:"shapes"`
}

type Candidate struct {
	ShapeType int
	X         float32
	Y         float32
	RX        float32
	RY        float32
	Theta     float32
	R         float32
	G         float32
	B         float32
	A         float32
}
