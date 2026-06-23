package gpu

import "forza-painter-geometrize-go/internal/model"

// ringSize is the number of candidate / result / grid host+device staging
// buffers kept in flight. With size 3 the engine can submit work for the
// next round while the previous round is still being read back, which is
// the foundation that lets us hide CPU candidate generation behind GPU
// kernels (e.g. apply + grid recompute) without ever stalling on a
// blocking transfer.
const ringSize = 3

// EvalResult holds the score and the optimal RGB color for a single
// evaluated candidate. RGB is computed analytically by the GPU; the engine
// stores it back into the candidate before applying the chosen shape.
type EvalResult struct {
	Score float32
	R     float32
	G     float32
	B     float32
}

// EvalTicket is returned by SubmitEval and consumed by WaitEval. It is
// just a typed handle into the evaluator's ring of in-flight slots; the
// sequence number guards against stale tickets being waited on twice.
type EvalTicket struct {
	slot  int
	seq   uint64
	count int
	valid bool
}

// GridTicket is the equivalent handle for an in-flight error-grid update.
type GridTicket struct {
	slot  int
	seq   uint64
	valid bool
}

// Valid reports whether the ticket refers to an actual in-flight
// submission. The zero value of GridTicket is invalid.
func (t GridTicket) Valid() bool { return t.valid }

// Valid reports whether the ticket refers to an actual in-flight
// submission. The zero value of EvalTicket is invalid.
func (t EvalTicket) Valid() bool { return t.valid }

// Backend abstracts a GPU compute backend (OpenCL or Vulkan).
type Backend interface {
	SubmitEval(cands []model.Candidate) (EvalTicket, error)
	WaitEval(t EvalTicket) ([]EvalResult, error)
	Evaluate(cands []model.Candidate) ([]EvalResult, error)
	SubmitApply(candidate model.Candidate) error
	Apply(candidate model.Candidate) error
	SubmitErrorGrid() (GridTicket, error)
	WaitErrorGrid(t GridTicket) ([]float32, int, int, error)
	ErrorGrid() ([]float32, int, int, error)
	ReadCurrent(dst []float32) error
	GridDims() (int, int)
	ImageDims() (int, int)
	ResetCurrentBuffer(current []float32) error
	Flush() error
	Close() error
	SetUseWorkGroupEval(bool)
	SetSampleStep(int)
	SetErrorMetric(string)
	SetSsimWeight(float32)
	SubmitSsimMap() error
}

// NewBackend creates the appropriate GPU backend based on name.
// Valid names are "opencl" (default) and "vulkan".
func NewBackend(name string, target, current []float32, mask []uint8, width, height, maxCandidates, gridSize int) (Backend, error) {
	switch name {
	case "vulkan":
		be, err := newVulkanBackend(target, current, mask, width, height, maxCandidates, gridSize)
		if err != nil {
			return nil, err
		}
		return be, nil
	default:
		return NewEvaluator(target, current, mask, width, height, maxCandidates, gridSize)
	}
}
