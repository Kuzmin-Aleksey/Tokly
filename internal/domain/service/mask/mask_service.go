package mask

import (
	"FairLAP/internal/domain/entity"
	"context"
	"fmt"
	"github.com/google/uuid"
	"image"
)

type RectRepo interface {
	GetRect(ctx context.Context, detectionId int) (*entity.RectDetection, string, error)
}

type PolygonRepo interface {
	GetPolygons(ctx context.Context, imageUid uuid.UUID) (*entity.Polygons, error)
}

type Service struct {
	rectRepo    RectRepo
	polygonRepo PolygonRepo
}

func NewService(rectRepo RectRepo, polygonRepo PolygonRepo) *Service {
	return &Service{
		rectRepo:    rectRepo,
		polygonRepo: polygonRepo,
	}
}

func (s *Service) GetRectMask(ctx context.Context, detectionId int) (image.Image, error) {
	const op = "service.GetRectMask"

	rectDetection, class, err := s.rectRepo.GetRect(ctx, detectionId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	mask := image.NewRGBA(rectDetection.ImgBounds())
	drawRectMask(mask, class, rectDetection)

	return mask, nil
}

func (s *Service) GetPolygonMask(ctx context.Context, imageUid uuid.UUID) (image.Image, error) {
	const op = "service.GetPolygonMask"

	p, err := s.polygonRepo.GetPolygons(ctx, imageUid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	mask := image.NewRGBA(p.ImgBounds())
	drawPolygon(mask, p.Data)

	return mask, nil
}
