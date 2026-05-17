package gpu

import (
	"fmt"
	"unsafe"

	"forza-painter-geometrize-go/internal/model"

	"github.com/jgillich/go-opencl/cl"
)

type Evaluator struct {
	context       *cl.Context
	queue         *cl.CommandQueue
	program       *cl.Program
	evalKernel    *cl.Kernel
	applyKernel   *cl.Kernel
	targetBuffer  *cl.MemObject
	currentBuffer *cl.MemObject
	maskBuffer    *cl.MemObject
	candBuffer    *cl.MemObject
	scoreBuffer   *cl.MemObject
	width         int
	height        int
	pixelCount    int
	maxCandidates int
	hostPacked    []float32
	hostScores    []float32
}

func NewEvaluator(target, current []float32, mask []uint8, width, height int, maxCandidates int) (*Evaluator, error) {
	if len(target) != len(current) {
		return nil, fmt.Errorf("target/current length mismatch")
	}
	if len(mask) != width*height {
		return nil, fmt.Errorf("mask length mismatch")
	}
	if maxCandidates < 1 {
		return nil, fmt.Errorf("maxCandidates must be > 0")
	}

	platforms, err := cl.GetPlatforms()
	if err != nil {
		return nil, err
	}
	if len(platforms) == 0 {
		return nil, fmt.Errorf("no OpenCL platforms found")
	}

	var device *cl.Device
	for _, p := range platforms {
		devices, dErr := p.GetDevices(cl.DeviceTypeGPU)
		if dErr == nil && len(devices) > 0 {
			device = devices[0]
			break
		}
	}
	if device == nil {
		for _, p := range platforms {
			devices, dErr := p.GetDevices(cl.DeviceTypeAll)
			if dErr == nil && len(devices) > 0 {
				device = devices[0]
				break
			}
		}
	}
	if device == nil {
		return nil, fmt.Errorf("no OpenCL device found")
	}

	ctx, err := cl.CreateContext([]*cl.Device{device})
	if err != nil {
		return nil, err
	}

	queue, err := ctx.CreateCommandQueue(device, 0)
	if err != nil {
		ctx.Release()
		return nil, err
	}

	program, err := ctx.CreateProgramWithSource([]string{evaluateKernelSource})
	if err != nil {
		queue.Release()
		ctx.Release()
		return nil, err
	}
	if err := program.BuildProgram([]*cl.Device{device}, "-cl-fast-relaxed-math -cl-mad-enable"); err != nil {
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, fmt.Errorf("failed building OpenCL program: %w", err)
	}

	evalKernel, err := program.CreateKernel("evaluate_candidates_v2")
	if err != nil {
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	applyKernel, err := program.CreateKernel("apply_candidate_v1")
	if err != nil {
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}

	targetBuffer, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, len(target)*4)
	if err != nil {
		applyKernel.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	currentBuffer, err := ctx.CreateEmptyBuffer(cl.MemReadWrite, len(current)*4)
	if err != nil {
		targetBuffer.Release()
		applyKernel.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	maskBuffer, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, len(mask))
	if err != nil {
		currentBuffer.Release()
		targetBuffer.Release()
		applyKernel.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	candBuffer, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, maxCandidates*9*4)
	if err != nil {
		maskBuffer.Release()
		currentBuffer.Release()
		targetBuffer.Release()
		applyKernel.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	scoreBuffer, err := ctx.CreateEmptyBuffer(cl.MemWriteOnly, maxCandidates*4)
	if err != nil {
		candBuffer.Release()
		maskBuffer.Release()
		currentBuffer.Release()
		targetBuffer.Release()
		applyKernel.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}

	if _, err := queue.EnqueueWriteBufferFloat32(targetBuffer, true, 0, target, nil); err != nil {
		return nil, err
	}
	if _, err := queue.EnqueueWriteBufferFloat32(currentBuffer, true, 0, current, nil); err != nil {
		return nil, err
	}
	if _, err := queue.EnqueueWriteBuffer(maskBuffer, true, 0, len(mask), unsafe.Pointer(&mask[0]), nil); err != nil {
		return nil, err
	}

	return &Evaluator{
		context:       ctx,
		queue:         queue,
		program:       program,
		evalKernel:    evalKernel,
		applyKernel:   applyKernel,
		targetBuffer:  targetBuffer,
		currentBuffer: currentBuffer,
		maskBuffer:    maskBuffer,
		candBuffer:    candBuffer,
		scoreBuffer:   scoreBuffer,
		width:         width,
		height:        height,
		pixelCount:    width * height,
		maxCandidates: maxCandidates,
		hostPacked:    make([]float32, maxCandidates*9),
		hostScores:    make([]float32, maxCandidates),
	}, nil
}

func (e *Evaluator) Close() {
	if e.scoreBuffer != nil {
		e.scoreBuffer.Release()
	}
	if e.candBuffer != nil {
		e.candBuffer.Release()
	}
	if e.maskBuffer != nil {
		e.maskBuffer.Release()
	}
	if e.currentBuffer != nil {
		e.currentBuffer.Release()
	}
	if e.targetBuffer != nil {
		e.targetBuffer.Release()
	}
	if e.applyKernel != nil {
		e.applyKernel.Release()
	}
	if e.evalKernel != nil {
		e.evalKernel.Release()
	}
	if e.program != nil {
		e.program.Release()
	}
	if e.queue != nil {
		e.queue.Release()
	}
	if e.context != nil {
		e.context.Release()
	}
}

func (e *Evaluator) Evaluate(candidates []model.Candidate) ([]float32, error) {
	count := len(candidates)
	if count == 0 {
		return nil, nil
	}
	if count > e.maxCandidates {
		return nil, fmt.Errorf("candidate count %d exceeds max %d", count, e.maxCandidates)
	}

	packed := e.hostPacked[:count*9]
	for i, c := range candidates {
		base := i * 9
		packed[base+0] = c.X
		packed[base+1] = c.Y
		packed[base+2] = c.RX
		packed[base+3] = c.RY
		packed[base+4] = c.Theta
		packed[base+5] = c.R
		packed[base+6] = c.G
		packed[base+7] = c.B
		packed[base+8] = c.A
	}

	if _, err := e.queue.EnqueueWriteBufferFloat32(e.candBuffer, true, 0, packed, nil); err != nil {
		return nil, err
	}

	if err := e.evalKernel.SetArgs(
		e.targetBuffer,
		e.currentBuffer,
		e.maskBuffer,
		e.candBuffer,
		e.scoreBuffer,
		int32(e.width),
		int32(e.height),
		int32(e.pixelCount),
	); err != nil {
		return nil, err
	}

	globalSize := []int{count}
	if _, err := e.queue.EnqueueNDRangeKernel(e.evalKernel, nil, globalSize, nil, nil); err != nil {
		return nil, err
	}

	scores := e.hostScores[:count]
	if _, err := e.queue.EnqueueReadBufferFloat32(e.scoreBuffer, true, 0, scores, nil); err != nil {
		return nil, err
	}
	out := make([]float32, count)
	copy(out, scores)
	return out, nil
}

func (e *Evaluator) Apply(candidate model.Candidate) error {
	if err := e.applyKernel.SetArgs(
		e.currentBuffer,
		e.maskBuffer,
		int32(e.width),
		int32(e.height),
		candidate.X,
		candidate.Y,
		candidate.RX,
		candidate.RY,
		candidate.Theta,
		candidate.R,
		candidate.G,
		candidate.B,
		candidate.A,
	); err != nil {
		return err
	}

	globalSize := []int{e.pixelCount}
	if _, err := e.queue.EnqueueNDRangeKernel(e.applyKernel, nil, globalSize, nil, nil); err != nil {
		return err
	}
	return nil
}

func (e *Evaluator) ReadCurrent(dst []float32) error {
	if len(dst) != e.pixelCount*4 {
		return fmt.Errorf("destination length mismatch")
	}
	_, err := e.queue.EnqueueReadBufferFloat32(e.currentBuffer, true, 0, dst, nil)
	return err
}
