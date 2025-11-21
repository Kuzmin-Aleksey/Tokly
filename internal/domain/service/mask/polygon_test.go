package mask

import (
	"image"
	"image/png"
	"os"
	"testing"
)

func TestDrawPolygon(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))

	coords := [][]image.Point{
		{
			{10, 10},
			{100, 1000},
			{10, 500},
			{10, 10},
		},
		{
			{500, 500},
			{900, 900},
			{10, 500},
			{500, 500},
		},
	}

	drawPolygon(img, coords)

	f, err := os.Create("D:\\polygon.png")
	if err != nil {
		t.Fatal(err)
	}

	png.Encode(f, img)

}
