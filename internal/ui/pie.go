package ui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// newPieRaster 生成自定义饼状图 (简单像素绘制)
func newPieRaster(labels []string, values []float64) fyne.CanvasObject {
	total := 0.0
	for _, v := range values {
		total += v
	}
	if total == 0 {
		total = 1
	}

	rand.Seed(time.Now().UnixNano())
	segColors := make([]color.NRGBA, len(values))
	for i := range segColors {
		// 生成偏亮但避免过白的颜色 (60~220)
		r := uint8(rand.Intn(160) + 60)
		g := uint8(rand.Intn(160) + 60)
		b := uint8(rand.Intn(160) + 60)
		segColors[i] = color.NRGBA{R: r, G: g, B: b, A: 255}
		fmt.Printf("[newPieRaster] segment %d color #%02x%02x%02x\n", i, r, g, b)
	}

	raster := canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		cx, cy := float64(w)/2, float64(h)/2
		dx, dy := float64(x)-cx, float64(y)-cy
		radius := math.Min(cx, cy) * 0.9
		if math.Hypot(dx, dy) > radius {
			return color.Transparent
		}
		ang := math.Atan2(dy, dx)
		if ang < 0 {
			ang += 2 * math.Pi
		}
		acc := 0.0
		for i, v := range values {
			frac := v / total
			if ang < (acc+frac)*2*math.Pi {
				return segColors[i]
			}
			acc += frac
		}
		return segColors[len(segColors)-1]
	})
	raster.SetMinSize(fyne.NewSize(300, 300))
	return raster
}
