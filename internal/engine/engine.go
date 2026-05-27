package engine

import (
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"time"

	"forza-painter-geometrize-go/internal/config"
	"forza-painter-geometrize-go/internal/gpu"
	"forza-painter-geometrize-go/internal/imageutil"
	"forza-painter-geometrize-go/internal/model"
	"forza-painter-geometrize-go/internal/output"
	"forza-painter-geometrize-go/internal/render"
)

type Options struct {
	ImagePath     string
	SettingsPath  string
	Profile       string
	OutputPath    string
	PreviewPath   string
	WorkspaceRoot string
	Seed          int64
	Backend       string
	ResumePath    string
}

const (
	maxNoImproveRetries = 100
	minImproveDelta     = -1e-7

	// Hill climb tuning. The mutation budget from settings is split into
	// up to maxHillClimbRounds rounds; each round mutates the current best
	// shape geometry slightly, evaluates the batch on GPU, and keeps any
	// improvement before starting the next round.
	maxHillClimbRounds  = 128
	idealHillClimbBatch = 1024
	minHillClimbRounds  = 1
)

func Run(opts Options) error {
	if opts.ImagePath == "" {
		return fmt.Errorf("image path is required")
	}
	if opts.WorkspaceRoot == "" {
		opts.WorkspaceRoot = "."
	}

	settingsPath, err := config.ResolveSettingsPath(opts.WorkspaceRoot, opts.SettingsPath, opts.Profile)
	if err != nil {
		return err
	}
	cfg, err := config.ParseSettings(settingsPath)
	if err != nil {
		return err
	}

	prepared, err := imageutil.LoadAndPrepare(opts.ImagePath, cfg.MaxResolution)
	if err != nil {
		return err
	}

	maxBatch := cfg.RandomSamples
	if cfg.MutatedSamples > maxBatch {
		maxBatch = cfg.MutatedSamples
	}
	gridSize := cfg.ErrorGridSize
	evaluator, err := gpu.NewBackend(opts.Backend, prepared.Target, prepared.Current, prepared.OpaqueMask, prepared.Width, prepared.Height, maxBatch, gridSize)
	if err != nil {
		return err
	}
	// fmt.Println("")

	// fmt.Println("=== Progressive Sampling ===")

	// fmt.Printf(
	// 	"Enabled: %v\n",
	// 	cfg.EnableProgressiveSampling,
	// )

	// fmt.Printf(
	// 	"Start Step: %d\n",
	// 	cfg.ProgressiveSamplingStart,
	// )

	// fmt.Printf(
	// 	"End Step: %d\n",
	// 	cfg.ProgressiveSamplingEnd,
	// )

	// fmt.Printf(
	// 	"Transition: %.3f\n",
	// 	cfg.ProgressiveSamplingTransition,
	// )

	// fmt.Printf(
	// 	"Curve: %.2f\n",
	// 	cfg.ProgressiveSamplingCurve,
	// )

	// maxReduction := cfg.ProgressiveSamplingStart *
	// 	cfg.ProgressiveSamplingStart

	// fmt.Printf(
	// 	"Max Pixel Reduction: 1/%d\n",
	// 	maxReduction,
	// )
	// if err != nil {
	// 	return err
	// }
	evaluator.SetUseWorkGroupEval(cfg.UseWorkGroupEval)
	defer evaluator.Close()

	fmt.Printf("Backend: %s\n", resolveBackendName(opts.Backend))

	rng := rand.New(rand.NewSource(seedValue(opts.Seed)))
	currentError, opaquePixels := computeTotalError(prepared.Target, prepared.Current, prepared.OpaqueMask)
	denom := float64(maxInt(1, opaquePixels*4))

	shapes := []model.Shape{backgroundShape(prepared, normalizeScore(currentError, denom))}
	acceptedShapes := 0

	moveStep, radiusStep := mutationSteps(prepared.Width, prepared.Height)
	hillClimbRounds, mutationsPerRound := planHillClimb(cfg.MutatedSamples)

	resumePath := opts.ResumePath
	if resumePath == "" {
		resumePath = cfg.LoadGeometry
	}
	if resumePath != "" {
		restoredShapes, restoredCount, resumeErr := restoreCheckpoint(resumePath, prepared, cfg.ForceOpaqueShapes, evaluator)
		if resumeErr != nil {
			return resumeErr
		}
		if restoredCount >= cfg.StopAt {
			return fmt.Errorf("checkpoint already has %d shapes (target stopAt=%d)", restoredCount, cfg.StopAt)
		}
		shapes = restoredShapes
		acceptedShapes = restoredCount
		currentError, opaquePixels = computeTotalError(prepared.Target, prepared.Current, prepared.OpaqueMask)
		denom = float64(maxInt(1, opaquePixels*4))
		fmt.Printf("Resumed from checkpoint: %s (%d/%d shapes)\n", resumePath, acceptedShapes, cfg.StopAt)
	}

	// Initial sampler is computed synchronously - the engine has nothing
	// useful to do until the first random batch can be sampled.
	initialGrid, gw, gh, err := evaluator.ErrorGrid()
	if err != nil {
		return err
	}
	sampler := newErrorSampler(initialGrid, gw, gh, prepared.Width, prepared.Height)
	var pendingGrid gpu.GridTicket // not valid initially

	fmt.Printf("Loaded image: %s (%dx%d), transparency=%v\n", opts.ImagePath, prepared.Width, prepared.Height, prepared.HasTransparency)
	fmt.Printf("Settings: stopAt=%d randomSamples=%d mutatedSamples=%d saveAt=%d saveEvery(preview)=%d\n",
		cfg.StopAt, cfg.RandomSamples, cfg.MutatedSamples, len(cfg.SaveAt), cfg.SaveEvery)
	fmt.Printf("Compatibility mode: forceOpaqueShapes=%v\n", cfg.ForceOpaqueShapes)
	fmt.Printf("Hill climb: %d rounds x %d mutations (move +/- %.1fpx, radius +/- %.1fpx, theta +/- 30deg)\n",
		hillClimbRounds, mutationsPerRound, moveStep, radiusStep)
	fmt.Println("Pipeline: async (in-order queue, ring=3; sampler 1-shape stale)")
	fmt.Println("Scoring mode: DeltaE with GPU-computed optimal color (negative = better)")

	consecutiveNoImprove := 0
	finalPruneAttempts := 0
	lastPrunedMilestone := 0
	const maxFinalPrunes = 5

	for acceptedShapes < cfg.StopAt {
		step := acceptedShapes + 1
		stepStart := time.Now()
		fmt.Printf("[%d/%d] Generating random samples (%d)...\n", step, cfg.StopAt, cfg.RandomSamples)
		// While we generate random candidates on the CPU, the GPU may
		// still be running the previous shape's apply + error-grid
		// kernels (queued non-blocking at the end of the last iteration).
		var progress float32
		if cfg.StopAt > 0 {
			progress = float32(acceptedShapes) / float32(cfg.StopAt)
		}

		evaluator.SetSampleStep(scoringSampleStep(cfg, progress))

		// fmt.Printf("[%d/%d] Scoring sample step: %d\n",
		// 	step, cfg.StopAt, evaluator.SampleStep)

		randomCands := randomCandidates(rng, prepared, cfg.RandomSamples, cfg.ForceOpaqueShapes, sampler, progress)

		fmt.Printf("[%d/%d] Evaluating random sample batch on GPU (%d)...\n", step, cfg.StopAt, len(randomCands))
		best, bestScore, err := submitAndPickBest(evaluator, randomCands)
		if err != nil {
			return err
		}
		fmt.Printf("[%d/%d] Random best delta: %.6f\n", step, cfg.StopAt, bestScore)

		if hillClimbRounds > 0 && mutationsPerRound > 0 && bestScore < 0 {
			improved := 0
			for round := 0; round < hillClimbRounds; round++ {
				mutations := mutatedCandidates(rng, prepared, best, mutationsPerRound, cfg.ForceOpaqueShapes, moveStep, radiusStep)
				roundBest, roundScore, mutErr := submitAndPickBest(evaluator, mutations)
				if mutErr != nil {
					return mutErr
				}
				if roundScore < bestScore {
					bestScore = roundScore
					best = roundBest
					improved++
				}
			}
			fmt.Printf("[%d/%d] Hill climb best delta after %d rounds: %.6f (%d improvement(s))\n",
				step, cfg.StopAt, hillClimbRounds, bestScore, improved)
		}

		if bestScore >= minImproveDelta {
			consecutiveNoImprove++
			fmt.Printf("[%d/%d] No improvement (delta %.6f). Retry %d/%d\n", step, cfg.StopAt, bestScore, consecutiveNoImprove, maxNoImproveRetries)
			if consecutiveNoImprove >= maxNoImproveRetries {
				fmt.Printf("Stopped early: reached max retries without improvement (%d)\n", maxNoImproveRetries)
				break
			}
			// Image state didn't change, sampler & pendingGrid still
			// describe the right canvas state; just loop and retry with
			// fresh random candidates.
			continue
		}

		consecutiveNoImprove = 0

		// Snap the accepted geometry onto the game's visible precision grid,
		// then re-evaluate that exact candidate so the applied canvas and
		// score bookkeeping match what the game will actually see.
		final := quantizeCandidate(best, prepared.Width, prepared.Height, cfg.ForceOpaqueShapes)
		final, finalScore, err := submitAndPickBest(evaluator, []model.Candidate{final})
		if err != nil {
			return err
		}
		if isRejectedEvalScore(finalScore) {
			consecutiveNoImprove++
			fmt.Printf("[%d/%d] Quantized candidate rejected after re-eval (delta %.6f). Retry %d/%d\n",
				step, cfg.StopAt, finalScore, consecutiveNoImprove, maxNoImproveRetries)
			if consecutiveNoImprove >= maxNoImproveRetries {
				fmt.Printf("Stopped early: reached max retries without improvement (%d)\n", maxNoImproveRetries)
				break
			}
			continue
		}

		// Submit apply non-blocking; the in-order queue ensures any
		// follow-up eval / grid kernel sees the updated canvas.
		if err := evaluator.SubmitApply(final); err != nil {
			return err
		}
		currentError += float64(finalScore)
		if currentError < 0 {
			currentError = 0
		}
		shapes = append(shapes, toShape(final, normalizeScore(currentError, denom)))
		acceptedShapes++
		fmt.Printf("[%d/%d] Added rotated ellipse #%d (delta %.6f)\n", acceptedShapes, cfg.StopAt, len(shapes)-1, finalScore)

		if shouldSave(acceptedShapes, cfg) {
			if err := saveShapes(opts, shapes, acceptedShapes); err != nil {
				return err
			}
			fmt.Printf("[%d/%d] Saved geometry checkpoint for shape count %d\n", acceptedShapes, cfg.StopAt, acceptedShapes)
		}

		if shouldSavePreview(acceptedShapes, cfg) {
			// ReadCurrent does a blocking read on the in-order queue, so
			// it implicitly waits for the apply we just submitted (and
			// any prior pending kernels). It does NOT touch the grid
			// ticket bookkeeping.
			if err := savePreviewSnapshot(evaluator, opts, prepared.Width, prepared.Height, acceptedShapes); err != nil {
				return err
			}
			if opts.PreviewPath != "" {
				fmt.Printf("[%d/%d] Saved preview snapshot\n", acceptedShapes, cfg.StopAt)
			}
		}

		isMilestonePass := acceptedShapes > 0 && acceptedShapes%500 == 0 && acceptedShapes > lastPrunedMilestone
		isFinalPass := acceptedShapes == cfg.StopAt && finalPruneAttempts < maxFinalPrunes

		if isMilestonePass || isFinalPass {
			if isFinalPass {
				fmt.Printf("[%d/%d] Reached target! Running final occlusion culling and compaction pass (%d/%d)...\n",
					acceptedShapes, cfg.StopAt, finalPruneAttempts+1, maxFinalPrunes)
			} else {
				fmt.Printf("[%d/%d] Scanning for completely occluded shapes to recycle...\n", acceptedShapes, cfg.StopAt)
			}

			if err := evaluator.Flush(); err != nil {
				return err
			}
			pendingGrid = gpu.GridTicket{} // invalidated by Flush()

			if isMilestonePass {
				lastPrunedMilestone = acceptedShapes
			}

			pruned := pruneOccludedShapes(shapes, prepared.Width, prepared.Height, prepared.OpaqueMask)
			removedCount := len(shapes) - len(pruned)

			if removedCount > 0 {
				if isFinalPass {
					fmt.Printf("[%d/%d] Recycled %d occluded shapes in final pass! Active shapes: %d -> %d\n",
						acceptedShapes, cfg.StopAt, removedCount, len(shapes), len(pruned))
					finalPruneAttempts++
				} else {
					fmt.Printf("[%d/%d] Recycled %d occluded shapes! Active shapes: %d -> %d\n",
						acceptedShapes, cfg.StopAt, removedCount, len(shapes), len(pruned))
				}
				shapes = pruned
				acceptedShapes = len(shapes) - 1

				if err := evaluator.ResetCurrentBuffer(prepared.Current); err != nil {
					return err
				}
				for _, s := range shapes[1:] {
					cand := model.Candidate{
						X:     float32(s.Data[0]),
						Y:     float32(s.Data[1]),
						RX:    float32(s.Data[2]),
						RY:    float32(s.Data[3]),
						Theta: float32(s.Data[4]),
						R:     float32(s.Color[0]) / 255.0,
						G:     float32(s.Color[1]) / 255.0,
						B:     float32(s.Color[2]) / 255.0,
						A:     float32(s.Color[3]) / 255.0,
					}
					if err := evaluator.SubmitApply(cand); err != nil {
						return err
					}
				}
				if err := evaluator.Flush(); err != nil {
					return err
				}
			} else {
				if isFinalPass {
					fmt.Printf("[%d/%d] Final pass: No occluded shapes found. Ready to finish.\n", acceptedShapes, cfg.StopAt)
				} else {
					fmt.Printf("[%d/%d] No occluded shapes found to recycle.\n", acceptedShapes, cfg.StopAt)
				}
			}
		}

		// Consume the previous shape's grid (its read finished long ago,
		// so this is essentially a free poll) and rebuild the sampler
		// from it. The sampler now reflects the canvas state one shape
		// behind real time; that's the cost of overlapping CPU random
		// generation with the GPU pipeline. Quality impact is negligible
		// because one shape changes <1% of pixels.
		if pendingGrid.Valid() {
			grid, gridW, gridH, gErr := evaluator.WaitErrorGrid(pendingGrid)
			if gErr != nil {
				return gErr
			}
			sampler = newErrorSampler(grid, gridW, gridH, prepared.Width, prepared.Height)
			pendingGrid = gpu.GridTicket{}
		}

		// Submit the grid kernel for the canvas-just-applied. It's
		// queued behind the apply; we'll consume the result next
		// iteration.
		newTicket, gErr := evaluator.SubmitErrorGrid()
		if gErr != nil {
			return gErr
		}
		pendingGrid = newTicket

		fmt.Printf("[%d/%d] Step completed in %s\n", acceptedShapes, cfg.StopAt, time.Since(stepStart).Round(time.Millisecond))
	}

	if acceptedShapes < cfg.StopAt {
		fmt.Printf("Finished early with %d/%d shapes due to no-improvement stopping rule\n", acceptedShapes, cfg.StopAt)
	}

	// Drain any pending grid ticket so its event is released cleanly.
	if pendingGrid.Valid() {
		if _, _, _, err := evaluator.WaitErrorGrid(pendingGrid); err != nil {
			return err
		}
		pendingGrid = gpu.GridTicket{}
	}

	if err := output.SaveGeometry(output.BuildFinalOutputPath(resolveOutputBase(opts)), shapes); err != nil {
		return err
	}

	if opts.PreviewPath != "" {
		current := make([]float32, prepared.Width*prepared.Height*4)
		if err := evaluator.ReadCurrent(current); err != nil {
			return err
		}
		if err := render.SavePNG(opts.PreviewPath, current, prepared.Width, prepared.Height); err != nil {
			return err
		}
	}

	return nil
}

func seedValue(seed int64) int64 {
	if seed != 0 {
		return seed
	}
	return time.Now().UnixNano()
}

func backgroundShape(p *imageutil.PreparedImage, score float64) model.Shape {
	return model.Shape{
		Type:  1,
		Data:  []int{0, 0, p.Width, p.Height},
		Color: []int{int(p.BackgroundRGBA[0]), int(p.BackgroundRGBA[1]), int(p.BackgroundRGBA[2]), int(p.BackgroundRGBA[3])},
		Score: score,
	}
}

// planHillClimb splits the configured mutation budget into a number of
// rounds and a per-round batch size. We aim for ~64 candidates per round
// to keep the GPU occupied while still giving the climb enough steps to
// walk uphill instead of just sampling around the random seed.
func planHillClimb(budget int) (rounds, perRound int) {
	if budget <= 0 {
		return 0, 0
	}
	rounds = budget / idealHillClimbBatch
	if rounds < minHillClimbRounds {
		rounds = minHillClimbRounds
	}
	if rounds > maxHillClimbRounds {
		rounds = maxHillClimbRounds
	}
	perRound = budget / rounds
	if perRound < 1 {
		perRound = 1
	}
	return rounds, perRound
}

func mutationSteps(width, height int) (move, radius float32) {
	diag := math.Sqrt(float64(width*width) + float64(height*height))
	move = float32(math.Max(2.0, diag*0.012))
	radius = float32(math.Max(2.0, diag*0.010))
	return move, radius
}

// errorSampler converts the GPU-produced error histogram into a CDF that
// can be sampled in O(log n) per draw. It is rebuilt every accepted shape.
type errorSampler struct {
	gridW, gridH int
	imgW, imgH   int
	cdf          []float64
	total        float64
}

func newErrorSampler(grid []float32, gridW, gridH, imgW, imgH int) *errorSampler {
	cdf := make([]float64, len(grid))
	var total float64
	for i, v := range grid {
		if v < 0 {
			v = 0
		}
		total += float64(v)
		cdf[i] = total
	}
	return &errorSampler{
		gridW: gridW,
		gridH: gridH,
		imgW:  imgW,
		imgH:  imgH,
		cdf:   cdf,
		total: total,
	}
}

func (s *errorSampler) sample(rng *rand.Rand) (float32, float32) {
	// Defensive nil-check first so the fallback below can safely deref s.
	if s == nil {
		return 0, 0
	}
	if s.total <= 0 || s.gridW <= 0 || s.gridH <= 0 {
		return rng.Float32() * float32(s.imgW), rng.Float32() * float32(s.imgH)
	}
	u := rng.Float64() * s.total
	lo, hi := 0, len(s.cdf)-1
	for lo < hi {
		mid := (lo + hi) / 2
		if s.cdf[mid] < u {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	cell := lo
	gx := cell % s.gridW
	gy := cell / s.gridW
	x0 := int(int64(gx) * int64(s.imgW) / int64(s.gridW))
	x1 := int(int64(gx+1) * int64(s.imgW) / int64(s.gridW))
	y0 := int(int64(gy) * int64(s.imgH) / int64(s.gridH))
	y1 := int(int64(gy+1) * int64(s.imgH) / int64(s.gridH))
	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}
	if x1 > s.imgW {
		x1 = s.imgW
	}
	if y1 > s.imgH {
		y1 = s.imgH
	}
	x := float32(x0) + rng.Float32()*float32(x1-x0)
	y := float32(y0) + rng.Float32()*float32(y1-y0)
	return x, y
}

// randomCandidates seeds candidates whose CENTER is biased towards the
// regions of the image that still have the most error. Geometry (radius,
// angle) is randomized; color is left zero because the GPU evaluator
// computes the optimal color analytically and writes it back in the
// EvalResult.
func randomCandidates(rng *rand.Rand, prepared *imageutil.PreparedImage, count int, forceOpaque bool, sampler *errorSampler, progress float32) []model.Candidate {
	out := make([]model.Candidate, 0, count)
	w := float32(prepared.Width)
	h := float32(prepared.Height)
	diag := float32(math.Sqrt(float64(prepared.Width*prepared.Width) + float64(prepared.Height*prepared.Height)))

	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	// Progressive scale decay: starts at 0.25 * diag and smoothly decays to 0.05 * diag.
	scaleFactor := float32(0.25 - 0.20*math.Pow(float64(progress), 1.5))
	maxRadius := diag * scaleFactor
	if maxRadius < 4 {
		maxRadius = 4
	}
	minRadius := float32(2)

	for i := 0; i < count; i++ {
		x, y := sampler.sample(rng)
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		if x > w-1 {
			x = w - 1
		}
		if y > h-1 {
			y = h - 1
		}
		alpha := float32(1.0)
		if !forceOpaque {
			alpha = randRange(rng, 0.3, 1.0)
		}
		out = append(out, quantizeCandidate(model.Candidate{
			X:     x,
			Y:     y,
			RX:    randRange(rng, minRadius, maxRadius),
			RY:    randRange(rng, minRadius, maxRadius),
			Theta: rng.Float32() * 360,
			A:     alpha,
		}, prepared.Width, prepared.Height, forceOpaque))
	}
	if len(out) == 0 {
		out = append(out, quantizeCandidate(model.Candidate{
			X:     w * 0.5,
			Y:     h * 0.5,
			RX:    maxRadius * 0.25,
			RY:    maxRadius * 0.25,
			Theta: 0,
			A:     1.0,
		}, prepared.Width, prepared.Height, forceOpaque))
	}
	return out
}

// mutatedCandidates only perturbs geometry. Colors are recomputed by the
// GPU on each evaluation, so seeding them on the CPU side would be wasted
// work (and would constrain the search).
func mutatedCandidates(rng *rand.Rand, prepared *imageutil.PreparedImage, base model.Candidate, count int, forceOpaque bool, moveStep, radiusStep float32) []model.Candidate {
	out := make([]model.Candidate, 0, count)
	w := float32(prepared.Width)
	h := float32(prepared.Height)
	for i := 0; i < count; i++ {
		cand := base
		cand.X += randRange(rng, -moveStep, moveStep)
		cand.Y += randRange(rng, -moveStep, moveStep)
		if cand.X < 0 {
			cand.X = 0
		}
		if cand.Y < 0 {
			cand.Y = 0
		}
		if cand.X > w-1 {
			cand.X = w - 1
		}
		if cand.Y > h-1 {
			cand.Y = h - 1
		}
		cand.RX = float32(math.Max(1, float64(cand.RX+randRange(rng, -radiusStep, radiusStep))))
		cand.RY = float32(math.Max(1, float64(cand.RY+randRange(rng, -radiusStep, radiusStep))))
		cand.Theta += randRange(rng, -30, 30)
		if cand.Theta < 0 {
			cand.Theta += 360
		}
		if cand.Theta >= 360 {
			cand.Theta -= 360
		}
		if forceOpaque {
			cand.A = 1.0
		}
		out = append(out, quantizeCandidate(cand, prepared.Width, prepared.Height, forceOpaque))
	}
	if len(out) == 0 {
		out = append(out, quantizeCandidate(base, prepared.Width, prepared.Height, forceOpaque))
	}
	return out
}

// submitAndPickBest submits a candidate batch, waits for the result and
// returns the lowest-score candidate with its GPU-computed optimal color
// merged in. This is the tight inner loop of both random sampling and
// hill climb.
func submitAndPickBest(e gpu.Backend, cands []model.Candidate) (model.Candidate, float32, error) {
	t, err := e.SubmitEval(cands)
	if err != nil {
		return model.Candidate{}, 0, err
	}
	results, err := e.WaitEval(t)
	if err != nil {
		return model.Candidate{}, 0, err
	}
	if len(results) == 0 {
		return model.Candidate{}, 0, fmt.Errorf("no candidate scores returned")
	}
	bestIdx := 0
	bestScore := results[0].Score
	for i := 1; i < len(results); i++ {
		if results[i].Score < bestScore {
			bestScore = results[i].Score
			bestIdx = i
		}
	}
	best := cands[bestIdx]
	best.R = results[bestIdx].R
	best.G = results[bestIdx].G
	best.B = results[bestIdx].B
	return best, bestScore, nil
}

func isRejectedEvalScore(score float32) bool {
	if math.IsNaN(float64(score)) || math.IsInf(float64(score), 0) {
		return true
	}
	return score >= float32(math.MaxFloat32)*0.5
}

func toShape(c model.Candidate, score float64) model.Shape {
	angle := int(math.Round(float64(c.Theta))) % 360
	if angle < 0 {
		angle += 360
	}
	if angle == 0 && c.Theta > 359.5 {
		angle = 360
	}
	return model.Shape{
		Type: 16,
		Data: []int{
			int(math.Round(float64(c.X))),
			int(math.Round(float64(c.Y))),
			maxInt(1, int(math.Round(float64(c.RX)))),
			maxInt(1, int(math.Round(float64(c.RY)))),
			angle,
		},
		Color: []int{int(f32ToByte(c.R)), int(f32ToByte(c.G)), int(f32ToByte(c.B)), int(f32ToByte(c.A))},
		Score: score,
	}
}

func f32ToByte(v float32) uint8 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return uint8(math.Round(float64(v * 255)))
}

func shouldSave(step int, cfg model.Settings) bool {
	_, ok := cfg.SaveAt[step]
	return ok
}

func shouldSavePreview(step int, cfg model.Settings) bool {
	if cfg.SaveEvery < 1 {
		return false
	}
	return step%cfg.SaveEvery == 0
}

func saveShapes(opts Options, shapes []model.Shape, step int) error {
	base := resolveOutputBase(opts)
	return output.SaveGeometry(output.BuildOutputPath(base, step), shapes)
}

func resolveOutputBase(opts Options) string {
	if opts.OutputPath != "" {
		return opts.OutputPath
	}
	ext := filepath.Ext(opts.ImagePath)
	if ext == "" {
		return opts.ImagePath
	}
	return opts.ImagePath
}

func randRange(rng *rand.Rand, minV, maxV float32) float32 {
	return minV + (maxV-minV)*rng.Float32()
}

func savePreviewSnapshot(evaluator gpu.Backend, opts Options, width, height, step int) error {
	if opts.PreviewPath == "" {
		return nil
	}
	ext := filepath.Ext(opts.PreviewPath)
	base := opts.PreviewPath
	if ext != "" {
		base = opts.PreviewPath[:len(opts.PreviewPath)-len(ext)]
	}
	outPath := fmt.Sprintf("%s.%d.png", base, step)
	current := make([]float32, width*height*4)
	if err := evaluator.ReadCurrent(current); err != nil {
		return err
	}
	return render.SavePNG(outPath, current, width, height)
}

func computeTotalError(target, current []float32, opaqueMask []uint8) (float64, int) {
	if len(target) != len(current) {
		return 0, 0
	}
	total := 0.0
	opaquePixels := 0
	for p := 0; p < len(opaqueMask); p++ {
		if opaqueMask[p] == 0 {
			continue
		}
		opaquePixels++
		idx := p * 4
		dr := float64(target[idx+0] - current[idx+0])
		dg := float64(target[idx+1] - current[idx+1])
		db := float64(target[idx+2] - current[idx+2])
		da := float64(target[idx+3] - current[idx+3])
		total += dr*dr + dg*dg + db*db + da*da
	}
	return total, opaquePixels
}

func normalizeScore(totalError, denom float64) float64 {
	if denom <= 0 {
		return 0
	}
	value := totalError / denom
	if value < 0 {
		value = 0
	}
	return math.Round(value*1_000_000) / 1_000_000
}

// quantizeCandidate snaps geometry onto the game's visible precision grid
// so the search evaluates the same values the game will actually consume.
// The JSON export remains integer-based for downstream tooling.
func quantizeCandidate(c model.Candidate, width, height int, forceOpaque bool) model.Candidate {
	c.X = snap2(clampFloat(c.X, 0, float32(maxInt(0, width-1))))
	c.Y = snap2(clampFloat(c.Y, 0, float32(maxInt(0, height-1))))
	c.RX = snap2(maxFloat(0.01, c.RX))
	c.RY = snap2(maxFloat(0.01, c.RY))
	c.Theta = snap2(normalizeAngle(c.Theta))

	if forceOpaque {
		c.A = 1.0
	} else {
		c.A = snap2(clampFloat(c.A, 0, 1))
	}
	c.R = snap2(clampFloat(c.R, 0, 1))
	c.G = snap2(clampFloat(c.G, 0, 1))
	c.B = snap2(clampFloat(c.B, 0, 1))
	return c
}

func snap2(v float32) float32 {
	return float32(math.Trunc(float64(v*100)) / 100)
}

func normalizeAngle(v float32) float32 {
	v = float32(math.Mod(float64(v), 360))
	if v < 0 {
		v += 360
	}
	return v
}

func clampFloat(v, minV, maxV float32) float32 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func maxFloat(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func resolveBackendName(name string) string {
	if name == "vulkan" {
		return "Vulkan"
	}
	return "OpenCL"
}

func scoringSampleStep(cfg model.Settings, progress float32) int {
	if !cfg.EnableProgressiveSampling {
		return 1
	}

	if progress >= cfg.ProgressiveSamplingTransition {
		return cfg.ProgressiveSamplingEnd
	}

	t := float64(progress / cfg.ProgressiveSamplingTransition)

	start := float64(cfg.ProgressiveSamplingStart)
	end := float64(cfg.ProgressiveSamplingEnd)
	curve := float64(cfg.ProgressiveSamplingCurve)

	step := end + (start-end)*math.Pow(1.0-t, curve)

	if step < 1 {
		return 1
	}

	return int(math.Round(step))
}

func pruneOccludedShapes(shapes []model.Shape, width, height int, opaqueMask []uint8) []model.Shape {
	if len(shapes) <= 1 {
		return shapes
	}

	// cov keeps track of pixels that are covered by 100% opaque shapes.
	// 0: not covered, 1: covered by an opaque shape
	cov := make([]uint8, width*height)

	// We iterate from the last shape down to the first shape (index 1).
	// Background shape at index 0 is always kept.
	keep := make([]bool, len(shapes))
	keep[0] = true // background shape is always kept

	for j := len(shapes) - 1; j >= 1; j-- {
		s := shapes[j]
		if s.Type != 16 {
			// If it's not a rotated ellipse, we just keep it.
			keep[j] = true
			continue
		}

		cx := float32(s.Data[0])
		cy := float32(s.Data[1])
		rx := float32(s.Data[2])
		ry := float32(s.Data[3])
		theta := float32(s.Data[4])
		alpha := s.Color[3] // 0-255

		if rx < 1 {
			rx = 1
		}
		if ry < 1 {
			ry = 1
		}

		t := theta * (math.Pi / 180.0)
		cosT := float32(math.Cos(float64(t)))
		sinT := float32(math.Sin(float64(t)))
		invRX2 := float32(1.0) / (rx * rx)
		invRY2 := float32(1.0) / (ry * ry)

		xMin := clampInt(int(cx-rx-1), 0, width-1)
		xMax := clampInt(int(cx+rx+1), 0, width-1)
		yMin := clampInt(int(cy-ry-1), 0, height-1)
		yMax := clampInt(int(cy+ry+1), 0, height-1)

		// Check if this shape is completely occluded by already-drawn opaque shapes
		isOccluded := true
		hasOpaquePixelsInsideMask := false

		for y := yMin; y <= yMax; y++ {
			for x := xMin; x <= xMax; x++ {
				p := y*width + x
				if opaqueMask[p] == 0 {
					continue
				}

				dx := float32(x) + 0.5 - cx
				dy := float32(y) + 0.5 - cy
				xr := dx*cosT + dy*sinT
				yr := -dx*sinT + dy*cosT
				if xr*xr*invRX2+yr*yr*invRY2 <= 1.0 {
					hasOpaquePixelsInsideMask = true
					if cov[p] == 0 {
						// This pixel of the ellipse is visible (not covered by any subsequent opaque shape)
						isOccluded = false
						break
					}
				}
			}
			if !isOccluded {
				break
			}
		}

		// If the shape covers no pixels inside the opaque mask, we can treat it as occluded/useless.
		if !hasOpaquePixelsInsideMask {
			isOccluded = true
		}

		if isOccluded {
			// This shape is completely covered, we don't keep it!
			keep[j] = false
		} else {
			keep[j] = true
			// If this shape is 100% opaque, mark all its pixels as covered in the cov mask
			if alpha == 255 {
				for y := yMin; y <= yMax; y++ {
					for x := xMin; x <= xMax; x++ {
						p := y*width + x
						if opaqueMask[p] == 0 {
							continue
						}
						dx := float32(x) + 0.5 - cx
						dy := float32(y) + 0.5 - cy
						xr := dx*cosT + dy*sinT
						yr := -dx*sinT + dy*cosT
						if xr*xr*invRX2+yr*yr*invRY2 <= 1.0 {
							cov[p] = 1
						}
					}
				}
			}
		}
	}

	// Rebuild the shapes list
	pruned := make([]model.Shape, 0, len(shapes))
	for j, k := range keep {
		if k {
			pruned = append(pruned, shapes[j])
		}
	}
	return pruned
}
