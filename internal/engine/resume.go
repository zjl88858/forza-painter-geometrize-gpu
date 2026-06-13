package engine

import (
	"encoding/json"
	"fmt"
	"os"

	"forza-painter-geometrize-go/internal/gpu"
	"forza-painter-geometrize-go/internal/imageutil"
	"forza-painter-geometrize-go/internal/model"
)

func restoreCheckpoint(path string, prepared *imageutil.PreparedImage, forceOpaque bool, evaluator gpu.Backend) ([]model.Shape, int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("read checkpoint: %w", err)
	}
	var payload model.Geometry
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, 0, fmt.Errorf("parse checkpoint json: %w", err)
	}
	if len(payload.Shapes) == 0 {
		return nil, 0, fmt.Errorf("checkpoint has no shapes")
	}

	currentError, opaquePixels := computeTotalError(prepared.Target, prepared.Current, prepared.OpaqueMask)
	denom := float64(maxInt(1, opaquePixels*4))
	shapes := []model.Shape{backgroundShape(prepared, normalizeScore(currentError, denom))}

	restored := 0
	for index, shape := range payload.Shapes {
		if index == 0 && looksLikeBackground(shape) {
			continue
		}
		candidate, ok := shapeToCandidate(shape)
		if !ok {
			continue
		}
		final := quantizeCandidate(candidate, prepared.Width, prepared.Height, forceOpaque)
		if err := evaluator.SubmitApply(final); err != nil {
			return nil, 0, err
		}
		shapes = append(shapes, toShape(final, shape.Score))
		restored++
	}
	if restored == 0 {
		return nil, 0, fmt.Errorf("checkpoint has no drawable rotated ellipses to restore")
	}
	if err := evaluator.Flush(); err != nil {
		return nil, 0, err
	}
	return shapes, restored, nil
}

func looksLikeBackground(shape model.Shape) bool {
	if shape.Type != 1 || len(shape.Data) < 4 || len(shape.Color) < 4 {
		return false
	}
	if shape.Color[3] <= 0 {
		return true
	}
	return shape.Data[2] >= 1 && shape.Data[3] >= 1
}

func shapeToCandidate(shape model.Shape) (model.Candidate, bool) {
	if (shape.Type != 0 && shape.Type != 16 && shape.Type != 2) || len(shape.Data) < 5 || len(shape.Color) < 4 {
		return model.Candidate{}, false
	}
	shapeType := 0
	if shape.Type == 16 {
		shapeType = 1
	} else if shape.Type == 2 {
		shapeType = 2
	}
	return model.Candidate{
		ShapeType: shapeType,
		X:     float32(shape.Data[0]),
		Y:     float32(shape.Data[1]),
		RX:    float32(shape.Data[2]),
		RY:    float32(shape.Data[3]),
		Theta: float32(shape.Data[4]),
		R:     float32(shape.Color[0]) / 255,
		G:     float32(shape.Color[1]) / 255,
		B:     float32(shape.Color[2]) / 255,
		A:     float32(shape.Color[3]) / 255,
	}, true
}
