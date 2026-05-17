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
}

const (
	maxNoImproveRetries = 100
	minImproveDelta     = -1e-7
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
	evaluator, err := gpu.NewEvaluator(prepared.Target, prepared.Current, prepared.OpaqueMask, prepared.Width, prepared.Height, maxBatch)
	if err != nil {
		return err
	}
	defer evaluator.Close()

	rng := rand.New(rand.NewSource(seedValue(opts.Seed)))
	currentError, opaquePixels := computeTotalError(prepared.Target, prepared.Current, prepared.OpaqueMask)
	denom := float64(maxInt(1, opaquePixels*4))

	shapes := []model.Shape{backgroundShape(prepared, normalizeScore(currentError, denom))}

	fmt.Printf("Loaded image: %s (%dx%d), transparency=%v\n", opts.ImagePath, prepared.Width, prepared.Height, prepared.HasTransparency)
	fmt.Printf("Settings: stopAt=%d randomSamples=%d mutatedSamples=%d saveAt=%d saveEvery(preview)=%d\n",
		cfg.StopAt, cfg.RandomSamples, cfg.MutatedSamples, len(cfg.SaveAt), cfg.SaveEvery)
	fmt.Println("Scoring mode: delta error (negative = better, positive = worse)")

	acceptedShapes := 0
	consecutiveNoImprove := 0

	for acceptedShapes < cfg.StopAt {
		step := acceptedShapes + 1
		stepStart := time.Now()
		fmt.Printf("[%d/%d] Generating random samples (%d)...\n", step, cfg.StopAt, cfg.RandomSamples)
		randomCands := randomCandidates(rng, prepared, cfg.RandomSamples)
		fmt.Printf("[%d/%d] Evaluating random sample batch on OpenCL (%d)...\n", step, cfg.StopAt, len(randomCands))
		best, bestScore, err := evaluateBest(evaluator, randomCands)
		if err != nil {
			return err
		}
		fmt.Printf("[%d/%d] Random best delta: %.6f\n", step, cfg.StopAt, bestScore)

		if cfg.MutatedSamples > 0 {
			fmt.Printf("[%d/%d] Generating mutated samples (%d)...\n", step, cfg.StopAt, cfg.MutatedSamples)
			mutations := mutatedCandidates(rng, prepared, best, cfg.MutatedSamples)
			fmt.Printf("[%d/%d] Evaluating mutated sample batch on OpenCL (%d)...\n", step, cfg.StopAt, len(mutations))
			mutBest, mutScore, mutErr := evaluateBest(evaluator, mutations)
			if mutErr != nil {
				return mutErr
			}
			if mutScore < bestScore {
				fmt.Printf("[%d/%d] Mutation improved delta: %.6f -> %.6f\n", step, cfg.StopAt, bestScore, mutScore)
				bestScore = mutScore
				best = mutBest
			} else {
				fmt.Printf("[%d/%d] Mutation did not improve delta (best remains %.6f)\n", step, cfg.StopAt, bestScore)
			}
		}

		if bestScore >= minImproveDelta {
			consecutiveNoImprove++
			fmt.Printf("[%d/%d] No improvement (delta %.6f). Retry %d/%d\n", step, cfg.StopAt, bestScore, consecutiveNoImprove, maxNoImproveRetries)
			if consecutiveNoImprove >= maxNoImproveRetries {
				fmt.Printf("Stopped early: reached max retries without improvement (%d)\n", maxNoImproveRetries)
				break
			}
			continue
		}

		consecutiveNoImprove = 0

		if err := evaluator.Apply(best); err != nil {
			return err
		}
		currentError += float64(bestScore)
		if currentError < 0 {
			currentError = 0
		}
		shapes = append(shapes, toShape(best, normalizeScore(currentError, denom)))
		acceptedShapes++
		fmt.Printf("[%d/%d] Added rotated ellipse #%d (delta %.6f)\n", acceptedShapes, cfg.StopAt, len(shapes)-1, bestScore)

		if shouldSave(acceptedShapes, cfg) {
			if err := saveShapes(opts, shapes, acceptedShapes); err != nil {
				return err
			}
			fmt.Printf("[%d/%d] Saved geometry checkpoint for shape count %d\n", acceptedShapes, cfg.StopAt, acceptedShapes)
		}

		if shouldSavePreview(acceptedShapes, cfg) {
			if err := savePreviewSnapshot(evaluator, opts, prepared.Width, prepared.Height, acceptedShapes); err != nil {
				return err
			}
			if opts.PreviewPath != "" {
				fmt.Printf("[%d/%d] Saved preview snapshot\n", acceptedShapes, cfg.StopAt)
			}
		}

		fmt.Printf("[%d/%d] Step completed in %s\n", acceptedShapes, cfg.StopAt, time.Since(stepStart).Round(time.Millisecond))
	}

	if acceptedShapes < cfg.StopAt {
		fmt.Printf("Finished early with %d/%d shapes due to no-improvement stopping rule\n", acceptedShapes, cfg.StopAt)
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

func randomCandidates(rng *rand.Rand, prepared *imageutil.PreparedImage, count int) []model.Candidate {
	out := make([]model.Candidate, 0, count)
	w := float32(prepared.Width)
	h := float32(prepared.Height)
	maxRadius := float32(math.Max(float64(prepared.Width), float64(prepared.Height))) * 0.35
	minRadius := float32(1)
	maxAttempts := count * 30
	attempts := 0

	for len(out) < count && attempts < maxAttempts {
		attempts++
		x := rng.Float32() * w
		y := rng.Float32() * h
		if prepared.OpaqueMask[int(y)*prepared.Width+int(x)] == 0 {
			continue
		}
		ci := (int(y)*prepared.Width + int(x)) * 4
		alpha := prepared.Target[ci+3]
		if alpha < 0.05 {
			continue
		}
		out = append(out, model.Candidate{
			X:     x,
			Y:     y,
			RX:    randRange(rng, minRadius, maxRadius),
			RY:    randRange(rng, minRadius, maxRadius),
			Theta: rng.Float32() * 360,
			R:     prepared.Target[ci+0],
			G:     prepared.Target[ci+1],
			B:     prepared.Target[ci+2],
			A:     randRange(rng, 0.2, 1.0) * alpha,
		})
	}
	if len(out) == 0 {
		out = append(out, model.Candidate{
			X:     w * 0.5,
			Y:     h * 0.5,
			RX:    maxRadius * 0.25,
			RY:    maxRadius * 0.25,
			Theta: 0,
			R:     0,
			G:     0,
			B:     0,
			A:     0,
		})
	}
	return out
}

func mutatedCandidates(rng *rand.Rand, prepared *imageutil.PreparedImage, base model.Candidate, count int) []model.Candidate {
	out := make([]model.Candidate, 0, count)
	maxAttempts := count * 40
	attempts := 0
	for len(out) < count && attempts < maxAttempts {
		attempts++
		cand := base
		cand.X += randRange(rng, -12, 12)
		cand.Y += randRange(rng, -12, 12)
		cand.RX = float32(math.Max(1, float64(cand.RX+randRange(rng, -8, 8))))
		cand.RY = float32(math.Max(1, float64(cand.RY+randRange(rng, -8, 8))))
		cand.Theta += randRange(rng, -30, 30)
		if cand.Theta < 0 {
			cand.Theta += 360
		}
		if cand.Theta >= 360 {
			cand.Theta -= 360
		}

		if cand.X < 0 || cand.Y < 0 || int(cand.X) >= prepared.Width || int(cand.Y) >= prepared.Height {
			continue
		}
		if prepared.OpaqueMask[int(cand.Y)*prepared.Width+int(cand.X)] == 0 {
			continue
		}

		ci := (int(cand.Y)*prepared.Width + int(cand.X)) * 4
		cand.R = clamp01(prepared.Target[ci+0] + randRange(rng, -0.08, 0.08))
		cand.G = clamp01(prepared.Target[ci+1] + randRange(rng, -0.08, 0.08))
		cand.B = clamp01(prepared.Target[ci+2] + randRange(rng, -0.08, 0.08))
		cand.A = clamp01(base.A + randRange(rng, -0.15, 0.15))
		out = append(out, cand)
	}
	if len(out) == 0 {
		out = append(out, base)
	}
	return out
}

func evaluateBest(e *gpu.Evaluator, cands []model.Candidate) (model.Candidate, float32, error) {
	scores, err := e.Evaluate(cands)
	if err != nil {
		return model.Candidate{}, 0, err
	}
	if len(scores) == 0 {
		return model.Candidate{}, 0, fmt.Errorf("no candidate scores returned")
	}
	bestIdx := 0
	bestScore := scores[0]
	for i := 1; i < len(scores); i++ {
		if scores[i] < bestScore {
			bestScore = scores[i]
			bestIdx = i
		}
	}
	return cands[bestIdx], bestScore, nil
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

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func savePreviewSnapshot(evaluator *gpu.Evaluator, opts Options, width, height, step int) error {
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
