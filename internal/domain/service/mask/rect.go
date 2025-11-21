package mask

import (
	"FairLAP/internal/domain/entity"
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
)

var opentypeFont, _ = opentype.Parse(fontTTF)
var colors = []color.RGBA{
	{0, 255, 0, 255},   // зеленый
	{255, 0, 0, 255},   // синий
	{0, 0, 255, 255},   // красный
	{255, 255, 0, 255}, // голубой
	{255, 0, 255, 255}, // пурпурный
	{255, 165, 0, 255}, // оранжевый
}

func drawRectMask(dst draw.Image, class string, rectDetection *entity.RectDetection) {
	rectBounds := rectDetection.Rect()

	fontSize := float64(rectDetection.ImgBounds().Dy()) / 40

	face, _ := opentype.NewFace(opentypeFont, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})

	thickness := max(int(float64(rectDetection.ImgBounds().Dy())/300), 1)

	c := colors[rectDetection.Id%len(colors)]

	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.White),
		Face: face,
	}

	fontHeight := drawer.Face.Metrics().Height.Floor()

	drawRect(dst, c, rectBounds, thickness)

	lbl := fmt.Sprintf("%s: %0.2f", class, rectDetection.Confidence)
	x, y := rectBounds.Min.X+thickness+2, rectBounds.Min.Y+fontHeight

	drawer.Dot = fixed.P(x, y)
	drawer.DrawString(lbl)
}

func drawRect(img draw.Image, c color.Color, r image.Rectangle, thickness int) {
	for j := range thickness {
		for i := r.Min.X; i <= r.Max.X; i++ {
			img.Set(i, r.Min.Y+j, c)
			img.Set(i, r.Max.Y+j, c)
		}
	}
	for j := range thickness {
		for i := r.Min.Y; i <= r.Max.Y; i++ {
			img.Set(r.Min.X+j, i, c)
			img.Set(r.Max.X+j, i, c)
		}
	}
}
