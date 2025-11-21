package mask

import (
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/color"
	"image/draw"
)

var blue = color.RGBA{0, 0, 255, 255}
var blueA = color.RGBA{0, 0, 255, 100}

func drawPolygon(dst draw.Image, polygons [][]image.Point) {
	ctx := draw2dimg.NewGraphicContext(dst)
	defer ctx.Close()

	ctx.SetFillColor(blueA)
	ctx.SetLineWidth(max(float64(dst.Bounds().Dy())/300, 1))
	ctx.SetStrokeColor(blue)

	for _, polygon := range polygons {

		if len(polygon) < 3 {
			continue
		}

		ctx.MoveTo(float64(polygon[0].X), float64(polygon[0].Y))

		polygon = append(polygon, polygon[0])[1:]

		for _, coord := range polygon {
			ctx.LineTo(float64(coord.X), float64(coord.Y))
		}

		ctx.FillStroke()
	}
}
