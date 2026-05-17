package render

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"forza-painter-geometrize-go/internal/model"
)

func ApplyEllipse(dst []float32, mask []uint8, width, height int, c model.Candidate) {
	if c.RX < 1 {
		c.RX = 1
	}
	if c.RY < 1 {
		c.RY = 1
	}
	t := c.Theta * (math.Pi / 180.0)
	cosT := float32(math.Cos(float64(t)))
	sinT := float32(math.Sin(float64(t)))
	invRX2 := float32(1.0) / (c.RX * c.RX)
	invRY2 := float32(1.0) / (c.RY * c.RY)

	xMin := clampInt(int(c.X-c.RX-1), 0, width-1)
	xMax := clampInt(int(c.X+c.RX+1), 0, width-1)
	yMin := clampInt(int(c.Y-c.RY-1), 0, height-1)
	yMax := clampInt(int(c.Y+c.RY+1), 0, height-1)

	for y := yMin; y <= yMax; y++ {
		for x := xMin; x <= xMax; x++ {
			if mask[y*width+x] == 0 {
				continue
			}
			dx := float32(x) + 0.5 - c.X
			dy := float32(y) + 0.5 - c.Y
			xr := dx*cosT + dy*sinT
			yr := -dx*sinT + dy*cosT
			inside := xr*xr*invRX2+yr*yr*invRY2 <= 1.0
			if !inside {
				continue
			}
			idx := (y*width + x) * 4
			alpha := c.A
			inv := 1.0 - alpha
			dst[idx+0] = dst[idx+0]*inv + c.R*alpha
			dst[idx+1] = dst[idx+1]*inv + c.G*alpha
			dst[idx+2] = dst[idx+2]*inv + c.B*alpha
			dst[idx+3] = dst[idx+3]*inv + alpha
		}
	}
}

func SavePNG(path string, pix []float32, width, height int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			img.SetRGBA(x, y, color.RGBA{
				R: toByte(pix[idx+0]),
				G: toByte(pix[idx+1]),
				B: toByte(pix[idx+2]),
				A: toByte(pix[idx+3]),
			})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func toByte(v float32) uint8 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return uint8(math.Round(float64(v * 255)))
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
