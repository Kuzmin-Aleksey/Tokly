package entity

import (
	"github.com/google/uuid"
	"image"
)

type Polygons struct {
	ImageUid uuid.UUID       `json:"image_uid"`
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Data     [][]image.Point `json:"Data"`
}

func (p Polygons) ImgBounds() image.Rectangle {
	return image.Rect(0, 0, p.Width, p.Height)
}
