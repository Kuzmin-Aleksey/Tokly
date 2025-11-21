package mask

import (
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/twpayne/go-geom"
	"image/color"
	"image/draw"
)

var blue = color.RGBA{0, 0, 255, 255}
var blueA = color.RGBA{0, 0, 255, 170}

func drawPolygon(dst draw.Image, polygons []geom.MultiPolygon) {
	ctx := draw2dimg.NewGraphicContext(dst)
	defer ctx.Close()

	ctx.SetFillColor(blueA)
	ctx.SetLineWidth(max(float64(dst.Bounds().Dy())/300, 1))
	ctx.SetStrokeColor(blue)

	for _, polygon := range polygons {

		coord0 := polygon.Coord(0)

		ctx.MoveTo(coord0.X(), coord0.Y())

		for i := range polygon.NumCoords() - 1 {
			coord := polygon.Coord(i)

			ctx.LineTo(coord.X(), coord.Y())
		}

		ctx.FillStroke()
	}

}
