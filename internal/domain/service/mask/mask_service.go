package mask

import (
	"FairLAP/internal/domain/entity"
	"context"
	"fmt"
	"github.com/google/uuid"
	"image"
	"image/color"
	"image/draw"
)

type Repo interface {
	GetRect(ctx context.Context, detectionId int) (*entity.RectDetection, string, error)
	GetPolygons(ctx context.Context, imageUid uuid.UUID) (*entity.Polygons, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetRectMask(ctx context.Context, detectionId int) (image.Image, error) {
	const op = "service.GetRectMask"

	rectDetection, class, err := s.repo.GetRect(ctx, detectionId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	mask := image.NewRGBA(rectDetection.ImgBounds())
	drawRectMask(mask, class, rectDetection)

	return mask, nil
}

func (s *Service) GetPolygonMask(ctx context.Context, imageUid uuid.UUID) (image.Image, error) {
	const op = "service.GetPolygonMask"

	p, err := s.repo.GetPolygons(ctx, imageUid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	mask := image.NewRGBA(p.ImgBounds())

	return mask, nil
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
