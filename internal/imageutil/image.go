package imageutil

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"golang.org/x/image/draw"
)

type PreparedImage struct {
	Width           int
	Height          int
	Target          []float32
	Current         []float32
	OpaqueMask      []uint8
	HasTransparency bool
	BackgroundRGBA  [4]uint8
}

func LoadAndPrepare(path string, maxResolution int) (*PreparedImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	scaled := resizeToMax(src, maxResolution)
	bounds := scaled.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid image size")
	}

	target := make([]float32, w*h*4)
	current := make([]float32, w*h*4)
	mask := make([]uint8, w*h)

	var sumR, sumG, sumB, sumA float64
	var opaqueCount float64
	hasTransparency := false

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r16, g16, b16, a16 := scaled.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			r := float32(r16) / 65535.0
			g := float32(g16) / 65535.0
			b := float32(b16) / 65535.0
			a := float32(a16) / 65535.0

			idx := (y*w + x) * 4
			target[idx+0] = r
			target[idx+1] = g
			target[idx+2] = b
			target[idx+3] = a

			if a > 0.001 {
				mask[y*w+x] = 1
				sumR += float64(r)
				sumG += float64(g)
				sumB += float64(b)
				sumA += float64(a)
				opaqueCount++
			} else {
				hasTransparency = true
			}
		}
	}

	bg := [4]uint8{0, 0, 0, 255}
	if opaqueCount > 0 {
		bg[0] = uint8(clamp255(sumR / opaqueCount * 255.0))
		bg[1] = uint8(clamp255(sumG / opaqueCount * 255.0))
		bg[2] = uint8(clamp255(sumB / opaqueCount * 255.0))
		bg[3] = uint8(clamp255(sumA / opaqueCount * 255.0))
	}
	if hasTransparency {
		bg[3] = 0
	}

	for i := 0; i < len(current); i += 4 {
		if hasTransparency {
			current[i+0] = 0
			current[i+1] = 0
			current[i+2] = 0
			current[i+3] = 0
		} else {
			current[i+0] = float32(bg[0]) / 255.0
			current[i+1] = float32(bg[1]) / 255.0
			current[i+2] = float32(bg[2]) / 255.0
			current[i+3] = 1.0
		}
	}

	return &PreparedImage{
		Width:           w,
		Height:          h,
		Target:          target,
		Current:         current,
		OpaqueMask:      mask,
		HasTransparency: hasTransparency,
		BackgroundRGBA:  bg,
	}, nil
}

func resizeToMax(src image.Image, maxResolution int) image.Image {
	if maxResolution <= 0 {
		return src
	}
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	maxDim := w
	if h > maxDim {
		maxDim = h
	}
	if maxDim <= maxResolution {
		return src
	}

	scale := float64(maxResolution) / float64(maxDim)
	nw := int(math.Round(float64(w) * scale))
	nh := int(math.Round(float64(h) * scale))
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

func clamp255(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
