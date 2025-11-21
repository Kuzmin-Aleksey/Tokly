package entity

import (
	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
	"image"
)

type Polygons struct {
	Id       int                `json:"id"`
	ImageUid uuid.UUID          `json:"image_uid"`
	Width    int                `json:"width"`
	Height   int                `json:"height"`
	Data     *geom.MultiPolygon `json:"Data"`
}

func (p Polygons) ImgBounds() image.Rectangle {
	return image.Rect(0, 0, p.Width, p.Height)
}

func (p Polygons) GetPoints() [][]image.Point {
	polygons := make([][]image.Point, p.Data.NumPolygons())

	for i := range polygons {
		coords := p.Data.Polygon(i).Coords()
		polygons[i] = make([]image.Point, len(coords))
		for j := range polygons[i] {
			polygons[i][j] = image.Pt(int(coords[j][0].X()), int(coords[j][0].Y()))
		}
	}

	return polygons
}
