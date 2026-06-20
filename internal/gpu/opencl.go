package gpu

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"forza-painter-geometrize-go/internal/model"

	"github.com/jgillich/go-opencl/cl"
)

// ErrorGridSize is the default side length of the downsampled error histogram.
// Overridable via settings.ini errorGridSize.
const DefaultErrorGridSize = 64

type evalSlot struct {
	readEvt *cl.Event
	seq     uint64
	busy    bool
}

type gridSlot struct {
	readEvt *cl.Event
	seq     uint64
	busy    bool
}

type Evaluator struct {
	context      *cl.Context
	queue        *cl.CommandQueue
	program      *cl.Program
	evalKernel   *cl.Kernel
	evalKernelV4 *cl.Kernel
	applyKernel  *cl.Kernel
	gridKernel   *cl.Kernel

	// SSIM kernels (created at init so compilation errors surface early).
	boxFilterHKernel    *cl.Kernel
	boxFilterVKernel    *cl.Kernel
	ssimEvalKernel      *cl.Kernel
	ssimErrorGridKernel *cl.Kernel

	UseWorkGroupEval bool

	SampleStep int

	// wgSize is the work-group size for evaluate_candidates_v4 (16×16).
	wgSize int

	targetBuffer  *cl.MemObject
	currentBuffer *cl.MemObject
	maskBuffer    *cl.MemObject

	// SSIM intermediate buffers for separable box filter.
	ssimMeanBufT  *cl.MemObject
	ssimMeanBufC  *cl.MemObject
	ssimMeanBufT2 *cl.MemObject
	ssimMeanBufC2 *cl.MemObject
	ssimMeanBufTC *cl.MemObject

	// Eval ring.
	candBuffers   [ringSize]*cl.MemObject
	resultBuffers [ringSize]*cl.MemObject
	hostCands     [ringSize][]float32
	hostResults   [ringSize][]float32
	evalSlots     [ringSize]evalSlot
	nextEvalSlot  int
	evalSeq       uint64

	// Grid ring.
	errorGridBufs  [ringSize]*cl.MemObject
	hostErrorGrids [ringSize][]float32
	gridSlots      [ringSize]gridSlot
	nextGridSlot   int
	gridSeq        uint64

	// SSIM map ring.
	ssimMapBufs     [ringSize]*cl.MemObject
	hostSsimMaps    [ringSize][]float32
	ssimMapSlots    [ringSize]gridSlot
	nextSsimMapSlot int
	ssimMapSeq      uint64

	// SSIM config.
	errorMetric     string  // "mse" (default) or "ssim"
	ssimWeight      float32 // blend weight; 0 = pure MSE, 1 = pure SSIM
	lastSsimMapSlot int     // most recently submitted SSIM map slot; -1 if none

	width         int
	height        int
	pixelCount    int
	maxCandidates int
	gridW         int
	gridH         int
}

func NewEvaluator(target, current []float32, mask []uint8, width, height, maxCandidates, gridSize int) (*Evaluator, error) {
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

	var allDevices []*cl.Device
	for _, p := range platforms {
		devices, dErr := p.GetDevices(cl.DeviceTypeGPU)
		if dErr == nil {
			allDevices = append(allDevices, devices...)
		}
	}
	if len(allDevices) == 0 {
		for _, p := range platforms {
			devices, dErr := p.GetDevices(cl.DeviceTypeAll)
			if dErr == nil {
				allDevices = append(allDevices, devices...)
			}
		}
	}
	if len(allDevices) == 0 {
		return nil, fmt.Errorf("no OpenCL devices found")
	}

	var device *cl.Device
	var bestScore int64 = -1
	for _, d := range allDevices {
		score := scoreDevice(d)
		if score > bestScore {
			bestScore = score
			device = d
		}
	}
	if device == nil {
		return nil, fmt.Errorf("no suitable OpenCL device found")
	}

	fmt.Printf("OpenCL: Selected device %q (Vendor: %q, GPU: %v, Discrete: %v, VRAM: %dMB, Compute Units: %d)\n",
		device.Name(),
		device.Vendor(),
		device.Type()&cl.DeviceTypeGPU != 0,
		!device.HostUnifiedMemory(),
		device.GlobalMemSize()/(1024*1024),
		device.MaxComputeUnits(),
	)

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

	evalKernel, err := program.CreateKernel("evaluate_candidates_v3")
	if err != nil {
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	evalKernelV4, err := program.CreateKernel("evaluate_candidates_v4")
	if err != nil {
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	applyKernel, err := program.CreateKernel("apply_candidate_v2")
	if err != nil {
		evalKernelV4.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}
	gridKernel, err := program.CreateKernel("compute_error_grid")
	if err != nil {
		applyKernel.Release()
		evalKernelV4.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, err
	}

	// SSIM kernels — pre-created so compilation errors surface at init.
	boxFilterHKernel, err := program.CreateKernel("box_filter_h")
	if err != nil {
		gridKernel.Release()
		applyKernel.Release()
		evalKernelV4.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, fmt.Errorf("box_filter_h kernel: %w", err)
	}
	boxFilterVKernel, err := program.CreateKernel("box_filter_v_ssim")
	if err != nil {
		boxFilterHKernel.Release()
		gridKernel.Release()
		applyKernel.Release()
		evalKernelV4.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, fmt.Errorf("box_filter_v_ssim kernel: %w", err)
	}
	ssimEvalKernel, err := program.CreateKernel("evaluate_candidates_ssim")
	if err != nil {
		boxFilterVKernel.Release()
		boxFilterHKernel.Release()
		gridKernel.Release()
		applyKernel.Release()
		evalKernelV4.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, fmt.Errorf("evaluate_candidates_ssim kernel: %w", err)
	}
	ssimErrorGridKernel, err := program.CreateKernel("compute_ssim_error_grid")
	if err != nil {
		ssimEvalKernel.Release()
		boxFilterVKernel.Release()
		boxFilterHKernel.Release()
		gridKernel.Release()
		applyKernel.Release()
		evalKernelV4.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
		return nil, fmt.Errorf("compute_ssim_error_grid kernel: %w", err)
	}

	gridW := gridSize
	gridH := gridSize
	if width < gridW {
		gridW = width
	}
	if height < gridH {
		gridH = height
	}
	if gridW < 1 {
		gridW = 1
	}
	if gridH < 1 {
		gridH = 1
	}

	e := &Evaluator{
		context:             ctx,
		queue:               queue,
		program:             program,
		evalKernel:          evalKernel,
		evalKernelV4:        evalKernelV4,
		applyKernel:         applyKernel,
		gridKernel:          gridKernel,
		boxFilterHKernel:    boxFilterHKernel,
		boxFilterVKernel:    boxFilterVKernel,
		ssimEvalKernel:      ssimEvalKernel,
		ssimErrorGridKernel: ssimErrorGridKernel,
		wgSize:              16, // 16×16 = 256 work-items per group
		width:               width,
		height:              height,
		pixelCount:          width * height,
		maxCandidates:       maxCandidates,
		gridW:               gridW,
		gridH:               gridH,
		SampleStep:          1,
		errorMetric:         "mse",
		ssimWeight:          0.5,
		lastSsimMapSlot:     -1,
	}

	cleanup := func() {
		for i := 0; i < ringSize; i++ {
			if e.candBuffers[i] != nil {
				e.candBuffers[i].Release()
			}
			if e.resultBuffers[i] != nil {
				e.resultBuffers[i].Release()
			}
			if e.errorGridBufs[i] != nil {
				e.errorGridBufs[i].Release()
			}
			if e.ssimMapBufs[i] != nil {
				e.ssimMapBufs[i].Release()
			}
		}
		if e.ssimMeanBufT != nil {
			e.ssimMeanBufT.Release()
		}
		if e.ssimMeanBufC != nil {
			e.ssimMeanBufC.Release()
		}
		if e.ssimMeanBufT2 != nil {
			e.ssimMeanBufT2.Release()
		}
		if e.ssimMeanBufC2 != nil {
			e.ssimMeanBufC2.Release()
		}
		if e.ssimMeanBufTC != nil {
			e.ssimMeanBufTC.Release()
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
		ssimErrorGridKernel.Release()
		ssimEvalKernel.Release()
		boxFilterVKernel.Release()
		boxFilterHKernel.Release()
		gridKernel.Release()
		applyKernel.Release()
		evalKernelV4.Release()
		evalKernel.Release()
		program.Release()
		queue.Release()
		ctx.Release()
	}

	if e.targetBuffer, err = ctx.CreateEmptyBuffer(cl.MemReadOnly, len(target)*4); err != nil {
		cleanup()
		return nil, err
	}
	if e.currentBuffer, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, len(current)*4); err != nil {
		cleanup()
		return nil, err
	}
	if e.maskBuffer, err = ctx.CreateEmptyBuffer(cl.MemReadOnly, len(mask)); err != nil {
		cleanup()
		return nil, err
	}

	for i := 0; i < ringSize; i++ {
		buf, bErr := ctx.CreateEmptyBuffer(cl.MemReadOnly, maxCandidates*6*4)
		if bErr != nil {
			cleanup()
			return nil, bErr
		}
		e.candBuffers[i] = buf
		rbuf, rErr := ctx.CreateEmptyBuffer(cl.MemWriteOnly, maxCandidates*4*4)
		if rErr != nil {
			cleanup()
			return nil, rErr
		}
		e.resultBuffers[i] = rbuf
		gbuf, gErr := ctx.CreateEmptyBuffer(cl.MemReadWrite, gridW*gridH*4)
		if gErr != nil {
			cleanup()
			return nil, gErr
		}
		e.errorGridBufs[i] = gbuf

		e.hostCands[i] = make([]float32, maxCandidates*6)
		e.hostResults[i] = make([]float32, maxCandidates*4)
		e.hostErrorGrids[i] = make([]float32, gridW*gridH)
	}

	// SSIM intermediate buffers — one per-pixel float each, needed for
	// the separable box-filter setup. Allocated lazily via ensureSsimBuffers.
	if e.ssimMeanBufT, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, e.pixelCount*4); err != nil {
		cleanup()
		return nil, err
	}
	if e.ssimMeanBufC, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, e.pixelCount*4); err != nil {
		cleanup()
		return nil, err
	}
	if e.ssimMeanBufT2, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, e.pixelCount*4); err != nil {
		cleanup()
		return nil, err
	}
	if e.ssimMeanBufC2, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, e.pixelCount*4); err != nil {
		cleanup()
		return nil, err
	}
	if e.ssimMeanBufTC, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, e.pixelCount*4); err != nil {
		cleanup()
		return nil, err
	}

	// SSIM map ring buffers — one per-pixel float per slot.
	for i := 0; i < ringSize; i++ {
		sbuf, sErr := ctx.CreateEmptyBuffer(cl.MemReadWrite, e.pixelCount*4)
		if sErr != nil {
			cleanup()
			return nil, sErr
		}
		e.ssimMapBufs[i] = sbuf
		e.hostSsimMaps[i] = make([]float32, e.pixelCount)
	}

	// Initial uploads. These are blocking because the engine has nothing
	// useful to do until the buffers are resident anyway. We still release
	// the returned events explicitly so the OpenCL runtime can free them
	// promptly instead of waiting for the Go finalizer to run.
	if evt, err := queue.EnqueueWriteBufferFloat32(e.targetBuffer, true, 0, target, nil); err != nil {
		cleanup()
		return nil, err
	} else if evt != nil {
		evt.Release()
	}
	if evt, err := queue.EnqueueWriteBufferFloat32(e.currentBuffer, true, 0, current, nil); err != nil {
		cleanup()
		return nil, err
	} else if evt != nil {
		evt.Release()
	}
	if evt, err := queue.EnqueueWriteBuffer(e.maskBuffer, true, 0, len(mask), unsafe.Pointer(&mask[0]), nil); err != nil {
		cleanup()
		return nil, err
	} else if evt != nil {
		evt.Release()
	}

	return e, nil
}

func (e *Evaluator) Close() error {
	_ = e.Flush()

	for i := 0; i < ringSize; i++ {
		if e.errorGridBufs[i] != nil {
			e.errorGridBufs[i].Release()
		}
		if e.resultBuffers[i] != nil {
			e.resultBuffers[i].Release()
		}
		if e.candBuffers[i] != nil {
			e.candBuffers[i].Release()
		}
		if e.ssimMapBufs[i] != nil {
			e.ssimMapBufs[i].Release()
		}
	}
	if e.ssimMeanBufT != nil {
		e.ssimMeanBufT.Release()
	}
	if e.ssimMeanBufC != nil {
		e.ssimMeanBufC.Release()
	}
	if e.ssimMeanBufT2 != nil {
		e.ssimMeanBufT2.Release()
	}
	if e.ssimMeanBufC2 != nil {
		e.ssimMeanBufC2.Release()
	}
	if e.ssimMeanBufTC != nil {
		e.ssimMeanBufTC.Release()
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
	if e.ssimErrorGridKernel != nil {
		e.ssimErrorGridKernel.Release()
	}
	if e.ssimEvalKernel != nil {
		e.ssimEvalKernel.Release()
	}
	if e.boxFilterVKernel != nil {
		e.boxFilterVKernel.Release()
	}
	if e.boxFilterHKernel != nil {
		e.boxFilterHKernel.Release()
	}
	if e.gridKernel != nil {
		e.gridKernel.Release()
	}
	if e.applyKernel != nil {
		e.applyKernel.Release()
	}
	if e.evalKernelV4 != nil {
		e.evalKernelV4.Release()
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
	return nil
}

// Flush blocks until every command previously enqueued on the internal
// command queue has finished and clears the bookkeeping for in-flight
// slots so they can be reused. Useful at the end of the run or before
// reading data through paths that assume no pending work (e.g. preview).
func (e *Evaluator) Flush() error {
	if e.queue == nil {
		return nil
	}
	if err := e.queue.Finish(); err != nil {
		return err
	}
	for i := 0; i < ringSize; i++ {
		if e.evalSlots[i].busy && e.evalSlots[i].readEvt != nil {
			e.evalSlots[i].readEvt.Release()
			e.evalSlots[i].readEvt = nil
			e.evalSlots[i].busy = false
		}
		if e.gridSlots[i].busy && e.gridSlots[i].readEvt != nil {
			e.gridSlots[i].readEvt.Release()
			e.gridSlots[i].readEvt = nil
			e.gridSlots[i].busy = false
		}
		if e.ssimMapSlots[i].busy && e.ssimMapSlots[i].readEvt != nil {
			e.ssimMapSlots[i].readEvt.Release()
			e.ssimMapSlots[i].readEvt = nil
			e.ssimMapSlots[i].busy = false
		}
	}
	return nil
}

// SubmitEval enqueues a candidate batch evaluation without blocking on the
// GPU. The caller must invoke WaitEval on the returned ticket before the
// host data backing the candidates becomes irrelevant; in practice the
// engine submits a batch, optionally does a tiny bit of CPU work, and then
// waits. The ring buffer guarantees that submitting more batches than the
// ring size is safe (the next slot is reclaimed with a defensive wait).
func (e *Evaluator) SubmitEval(cands []model.Candidate) (EvalTicket, error) {
	count := len(cands)
	if count == 0 {
		return EvalTicket{}, nil
	}
	if count > e.maxCandidates {
		return EvalTicket{}, fmt.Errorf("candidate count %d exceeds max %d", count, e.maxCandidates)
	}

	slot := e.nextEvalSlot
	e.nextEvalSlot = (e.nextEvalSlot + 1) % ringSize

	if e.evalSlots[slot].busy {
		// Engine forgot to consume an outstanding ticket; reclaim the slot.
		err := cl.WaitForEvents([]*cl.Event{e.evalSlots[slot].readEvt})
		e.evalSlots[slot].readEvt.Release()
		e.evalSlots[slot].readEvt = nil
		e.evalSlots[slot].busy = false
		if err != nil {
			return EvalTicket{}, err
		}
	}

	packed := e.hostCands[slot][:count*6]
	for i, c := range cands {
		base := i * 6
		packed[base+0] = c.X
		packed[base+1] = c.Y
		packed[base+2] = c.RX
		packed[base+3] = c.RY
		packed[base+4] = c.Theta
		packed[base+5] = c.A
	}

	// Non-blocking write. The OpenCL spec snapshots the host pointer's
	// contents at command submission, but the runtime is allowed to defer
	// the actual transfer; either way our hostCands slot is not reused
	// until the ring wraps and reclaims it.
	writeEvt, err := e.queue.EnqueueWriteBufferFloat32(e.candBuffers[slot], false, 0, packed, nil)
	if err != nil {
		return EvalTicket{}, err
	}
	writeEvt.Release()

	if e.UseWorkGroupEval {
		if err := e.evalKernelV4.SetArgs(
			e.targetBuffer,
			e.currentBuffer,
			e.maskBuffer,
			e.candBuffers[slot],
			e.resultBuffers[slot],
			int32(e.width),
			int32(e.height),
			int32(e.SampleStep),
		); err != nil {
			return EvalTicket{}, err
		}
		gs := e.wgSize
		kernelEvt, err := e.queue.EnqueueNDRangeKernel(
			e.evalKernelV4, nil,
			[]int{count * gs, gs},
			[]int{gs, gs},
			nil,
		)
		if err != nil {
			return EvalTicket{}, err
		}
		kernelEvt.Release()
	} else if e.errorMetric == "ssim" {
		// SSIM-aware evaluation: reads from the SSIM error map.
		ssimBuf := e.ssimMapBufs[0]
		if e.lastSsimMapSlot >= 0 {
			ssimBuf = e.ssimMapBufs[e.lastSsimMapSlot]
		}
		if err := e.ssimEvalKernel.SetArgs(
			e.targetBuffer,
			e.currentBuffer,
			e.maskBuffer,
			e.candBuffers[slot],
			ssimBuf,
			e.resultBuffers[slot],
			int32(e.width),
			int32(e.height),
			int32(e.SampleStep),
			e.ssimWeight,
		); err != nil {
			return EvalTicket{}, err
		}
		kernelEvt, err := e.queue.EnqueueNDRangeKernel(e.ssimEvalKernel, nil, []int{count}, nil, nil)
		if err != nil {
			return EvalTicket{}, err
		}
		kernelEvt.Release()
	} else {
		if err := e.evalKernel.SetArgs(
			e.targetBuffer,
			e.currentBuffer,
			e.maskBuffer,
			e.candBuffers[slot],
			e.resultBuffers[slot],
			int32(e.width),
			int32(e.height),
			int32(e.SampleStep),
		); err != nil {
			return EvalTicket{}, err
		}
		kernelEvt, err := e.queue.EnqueueNDRangeKernel(e.evalKernel, nil, []int{count}, nil, nil)
		if err != nil {
			return EvalTicket{}, err
		}
		kernelEvt.Release()
	}

	flat := e.hostResults[slot][:count*4]
	readEvt, err := e.queue.EnqueueReadBufferFloat32(e.resultBuffers[slot], false, 0, flat, nil)
	if err != nil {
		return EvalTicket{}, err
	}

	e.evalSeq++
	e.evalSlots[slot] = evalSlot{
		readEvt: readEvt,
		seq:     e.evalSeq,
		busy:    true,
	}
	return EvalTicket{slot: slot, seq: e.evalSeq, count: count, valid: true}, nil
}

// WaitEval blocks until the given submission's read transfer completes and
// returns the per-candidate results. After this call the slot is freed and
// the ticket is exhausted.
func (e *Evaluator) WaitEval(t EvalTicket) ([]EvalResult, error) {
	if !t.valid {
		return nil, nil
	}
	s := &e.evalSlots[t.slot]
	if !s.busy || s.seq != t.seq {
		return nil, fmt.Errorf("WaitEval: stale or invalid ticket")
	}
	err := cl.WaitForEvents([]*cl.Event{s.readEvt})
	s.readEvt.Release()
	s.readEvt = nil
	s.busy = false
	if err != nil {
		return nil, err
	}

	flat := e.hostResults[t.slot][:t.count*4]
	out := make([]EvalResult, t.count)
	for i := 0; i < t.count; i++ {
		out[i] = EvalResult{
			Score: flat[i*4+0],
			R:     flat[i*4+1],
			G:     flat[i*4+2],
			B:     flat[i*4+3],
		}
	}
	return out, nil
}

// Evaluate is the synchronous helper used by code paths that don't care
// about overlap. It is implemented in terms of Submit/Wait.
func (e *Evaluator) Evaluate(cands []model.Candidate) ([]EvalResult, error) {
	t, err := e.SubmitEval(cands)
	if err != nil {
		return nil, err
	}
	return e.WaitEval(t)
}

// SubmitApply enqueues a blend kernel for the given candidate without
// blocking. The in-order command queue guarantees that any subsequent
// SubmitEval / SubmitErrorGrid will observe the updated current canvas.
func (e *Evaluator) SubmitApply(candidate model.Candidate) error {
	rx := candidate.RX
	ry := candidate.RY
	if rx < 1 {
		rx = 1
	}
	if ry < 1 {
		ry = 1
	}
	theta := float64(candidate.Theta) * (math.Pi / 180.0)
	cosT := math.Cos(theta)
	sinT := math.Sin(theta)
	rx2 := float64(rx) * float64(rx)
	ry2 := float64(ry) * float64(ry)
	cos2 := cosT * cosT
	sin2 := sinT * sinT
	ex := math.Sqrt(rx2*cos2 + ry2*sin2)
	ey := math.Sqrt(rx2*sin2 + ry2*cos2)

	xMin := int(math.Floor(float64(candidate.X) - ex - 1.0))
	xMax := int(math.Ceil(float64(candidate.X) + ex + 1.0))
	yMin := int(math.Floor(float64(candidate.Y) - ey - 1.0))
	yMax := int(math.Ceil(float64(candidate.Y) + ey + 1.0))
	if xMin < 0 {
		xMin = 0
	}
	if yMin < 0 {
		yMin = 0
	}
	if xMax > e.width-1 {
		xMax = e.width - 1
	}
	if yMax > e.height-1 {
		yMax = e.height - 1
	}
	if xMax < xMin || yMax < yMin {
		return nil
	}

	bw := xMax - xMin + 1
	bh := yMax - yMin + 1

	if err := e.applyKernel.SetArgs(
		e.currentBuffer,
		e.maskBuffer,
		int32(e.width),
		int32(e.height),
		int32(xMin),
		int32(yMin),
		int32(xMax),
		int32(yMax),
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

	evt, err := e.queue.EnqueueNDRangeKernel(e.applyKernel, nil, []int{bw, bh}, nil, nil)
	if err != nil {
		return err
	}
	evt.Release()
	return nil
}

// Apply is the synchronous helper used by tests / single-call flows. It
// drains the queue via Flush so that any in-flight ticket events that
// happen to complete during the wait are also reaped, keeping the slot
// bookkeeping consistent with the async path.
func (e *Evaluator) Apply(candidate model.Candidate) error {
	if err := e.SubmitApply(candidate); err != nil {
		return err
	}
	return e.Flush()
}

// SubmitErrorGrid enqueues a downsampled error histogram computation and
// the corresponding readback without blocking. The histogram cells reflect
// the canvas state at the moment the kernel runs in the queue, i.e. after
// any apply commands previously submitted on the same queue.
func (e *Evaluator) SubmitErrorGrid() (GridTicket, error) {
	slot := e.nextGridSlot
	e.nextGridSlot = (e.nextGridSlot + 1) % ringSize

	if e.gridSlots[slot].busy {
		err := cl.WaitForEvents([]*cl.Event{e.gridSlots[slot].readEvt})
		e.gridSlots[slot].readEvt.Release()
		e.gridSlots[slot].readEvt = nil
		e.gridSlots[slot].busy = false
		if err != nil {
			return GridTicket{}, err
		}
	}

	if e.errorMetric == "ssim" {
		// SSIM error grid: samples from the pre-computed SSIM map.
		ssimBuf := e.ssimMapBufs[0]
		if e.lastSsimMapSlot >= 0 {
			ssimBuf = e.ssimMapBufs[e.lastSsimMapSlot]
		}
		if err := e.ssimErrorGridKernel.SetArgs(
			ssimBuf,
			e.maskBuffer,
			e.errorGridBufs[slot],
			int32(e.width),
			int32(e.height),
			int32(e.gridW),
			int32(e.gridH),
		); err != nil {
			return GridTicket{}, err
		}
	} else {
		if err := e.gridKernel.SetArgs(
			e.targetBuffer,
			e.currentBuffer,
			e.maskBuffer,
			e.errorGridBufs[slot],
			int32(e.width),
			int32(e.height),
			int32(e.gridW),
			int32(e.gridH),
		); err != nil {
			return GridTicket{}, err
		}
	}
	kernelEvt, err := e.queue.EnqueueNDRangeKernel(e.gridKernel, nil, []int{e.gridW, e.gridH}, nil, nil)
	if err != nil {
		return GridTicket{}, err
	}
	kernelEvt.Release()

	readEvt, err := e.queue.EnqueueReadBufferFloat32(e.errorGridBufs[slot], false, 0, e.hostErrorGrids[slot], nil)
	if err != nil {
		return GridTicket{}, err
	}

	e.gridSeq++
	e.gridSlots[slot] = gridSlot{
		readEvt: readEvt,
		seq:     e.gridSeq,
		busy:    true,
	}
	return GridTicket{slot: slot, seq: e.gridSeq, valid: true}, nil
}

// WaitErrorGrid blocks until the given submission's grid readback finishes
// and returns a copy of the grid plus its dimensions.
func (e *Evaluator) WaitErrorGrid(t GridTicket) ([]float32, int, int, error) {
	if !t.valid {
		return nil, 0, 0, fmt.Errorf("WaitErrorGrid: invalid ticket")
	}
	s := &e.gridSlots[t.slot]
	if !s.busy || s.seq != t.seq {
		return nil, 0, 0, fmt.Errorf("WaitErrorGrid: stale or invalid ticket")
	}
	err := cl.WaitForEvents([]*cl.Event{s.readEvt})
	s.readEvt.Release()
	s.readEvt = nil
	s.busy = false
	if err != nil {
		return nil, 0, 0, err
	}
	out := make([]float32, len(e.hostErrorGrids[t.slot]))
	copy(out, e.hostErrorGrids[t.slot])
	return out, e.gridW, e.gridH, nil
}

// ErrorGrid is the synchronous helper retained for callers that don't
// care about overlap.
func (e *Evaluator) ErrorGrid() ([]float32, int, int, error) {
	t, err := e.SubmitErrorGrid()
	if err != nil {
		return nil, 0, 0, err
	}
	return e.WaitErrorGrid(t)
}

// ReadCurrent reads the entire current canvas back to host memory. The
// blocking read goes through the same in-order queue as everything else,
// so it serializes naturally after any pending apply / grid commands
// without us having to drain the queue (which would invalidate any
// outstanding GridTicket the caller is still holding).
func (e *Evaluator) ReadCurrent(dst []float32) error {
	if len(dst) != e.pixelCount*4 {
		return fmt.Errorf("destination length mismatch")
	}
	evt, err := e.queue.EnqueueReadBufferFloat32(e.currentBuffer, true, 0, dst, nil)
	if evt != nil {
		evt.Release()
	}
	return err
}

// GridDims returns the dimensions of the error histogram grid.
func (e *Evaluator) GridDims() (int, int) {
	return e.gridW, e.gridH
}

// ImageDims returns the working image dimensions.
func (e *Evaluator) ImageDims() (int, int) {
	return e.width, e.height
}

// ResetCurrentBuffer clears/re-uploads the initial current canvas back to e.currentBuffer.
func (e *Evaluator) ResetCurrentBuffer(current []float32) error {
	if len(current) != e.pixelCount*4 {
		return fmt.Errorf("current slice length mismatch")
	}
	evt, err := e.queue.EnqueueWriteBufferFloat32(e.currentBuffer, true, 0, current, nil)
	if evt != nil {
		evt.Release()
	}
	return err
}

func (e *Evaluator) SetUseWorkGroupEval(v bool) { e.UseWorkGroupEval = v }
func (e *Evaluator) SetSampleStep(v int)        { e.SampleStep = v }

// SetErrorMetric configures the error metric used by the evaluator.
// Valid values are "mse" (default) and "ssim".
func (e *Evaluator) SetErrorMetric(metric string) {
	metric = strings.ToLower(strings.TrimSpace(metric))
	if metric != "mse" && metric != "ssim" {
		return
	}
	e.errorMetric = metric
}

// SetSsimWeight sets the blend weight for SSIM mode.
// 0 = pure MSE, 1 = pure SSIM (default 0.5).
func (e *Evaluator) SetSsimWeight(w float32) {
	if w < 0 {
		w = 0
	}
	if w > 1 {
		w = 1
	}
	e.ssimWeight = w
}

// SubmitSsimMap enqueues the two-pass box-filter → SSIM map computation
// without blocking. The returned GridTicket can be waited on with
// WaitErrorGrid (reuses the same ticket type since the readback flow
// is identical).
func (e *Evaluator) SubmitSsimMap() (GridTicket, error) {
	slot := e.nextSsimMapSlot
	e.nextSsimMapSlot = (e.nextSsimMapSlot + 1) % ringSize

	if e.ssimMapSlots[slot].busy {
		err := cl.WaitForEvents([]*cl.Event{e.ssimMapSlots[slot].readEvt})
		e.ssimMapSlots[slot].readEvt.Release()
		e.ssimMapSlots[slot].readEvt = nil
		e.ssimMapSlots[slot].busy = false
		if err != nil {
			return GridTicket{}, err
		}
	}

	// Pass 1: horizontal box filter — target/current → 5 mean buffers.
	if err := e.boxFilterHKernel.SetArgs(
		e.targetBuffer,
		e.currentBuffer,
		e.ssimMeanBufT,
		e.ssimMeanBufC,
		e.ssimMeanBufT2,
		e.ssimMeanBufC2,
		e.ssimMeanBufTC,
		int32(e.width),
		int32(e.height),
	); err != nil {
		return GridTicket{}, err
	}
	if evt, err := e.queue.EnqueueNDRangeKernel(e.boxFilterHKernel, nil, []int{e.width, e.height}, nil, nil); err != nil {
		return GridTicket{}, err
	} else {
		evt.Release()
	}

	// Pass 2: vertical box filter + SSIM → ssim_map buffer.
	if err := e.boxFilterVKernel.SetArgs(
		e.ssimMeanBufT,
		e.ssimMeanBufC,
		e.ssimMeanBufT2,
		e.ssimMeanBufC2,
		e.ssimMeanBufTC,
		e.ssimMapBufs[slot],
		int32(e.width),
		int32(e.height),
	); err != nil {
		return GridTicket{}, err
	}
	if evt, err := e.queue.EnqueueNDRangeKernel(e.boxFilterVKernel, nil, []int{e.width, e.height}, nil, nil); err != nil {
		return GridTicket{}, err
	} else {
		evt.Release()
	}

	readEvt, err := e.queue.EnqueueReadBufferFloat32(e.ssimMapBufs[slot], false, 0, e.hostSsimMaps[slot], nil)
	if err != nil {
		return GridTicket{}, err
	}

	e.ssimMapSeq++
	e.ssimMapSlots[slot] = gridSlot{
		readEvt: readEvt,
		seq:     e.ssimMapSeq,
		busy:    true,
	}
	e.lastSsimMapSlot = slot
	return GridTicket{slot: slot, seq: e.ssimMapSeq, valid: true}, nil
}

func scoreDevice(d *cl.Device) int64 {
	var score int64

	// 1. Prioritize GPUs over CPUs or other device types
	if d.Type()&cl.DeviceTypeGPU != 0 {
		score += 1_000_000_000_000
	}

	// 2. Prioritize Discrete GPUs (HostUnifiedMemory == false) over Integrated GPUs (HostUnifiedMemory == true)
	if !d.HostUnifiedMemory() {
		score += 500_000_000_000
	}

	// 3. Size of Global Memory (VRAM) as factor
	memMB := d.GlobalMemSize() / (1024 * 1024)
	score += memMB * 10_000

	// 4. Number of parallel compute units
	score += int64(d.MaxComputeUnits()) * 10_000

	// 5. Vendor/Hardware heuristics
	vendor := strings.ToLower(d.Vendor())
	name := strings.ToLower(d.Name())
	if strings.Contains(vendor, "nvidia") || strings.Contains(name, "geforce") || strings.Contains(name, "rtx") {
		score += 5_000_000_000
	} else if strings.Contains(vendor, "amd") || strings.Contains(name, "radeon") {
		score += 3_000_000_000
	} else if strings.Contains(vendor, "intel") {
		score += 1_000_000_000
	}

	return score
}
