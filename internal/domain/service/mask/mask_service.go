package mask

import (
	"FairLAP/internal/domain/entity"
	"context"
	"fmt"
	"github.com/google/uuid"
	"image"
	"image/jpeg"
	"os"
	"time"
)

type RectRepo interface {
	GetRect(ctx context.Context, detectionId int) (*entity.RectDetection, string, error)
}

type Polygons interface {
	DetectPolygons(img image.Image) ([][]image.Point, error)
}

type Images interface {
	Open(groupId int, uid uuid.UUID) (*os.File, error)
}

type Service struct {
	rectRepo RectRepo
	polygons Polygons
	images   Images

	polygonCache map[uuid.UUID]cachedImage
}

type cachedImage struct {
	img image.Image
	ts  time.Time
}

func NewService(rectRepo RectRepo, polygons Polygons, images Images) *Service {
	s := &Service{
		rectRepo: rectRepo,
		polygons: polygons,
		images:   images,

		polygonCache: make(map[uuid.UUID]cachedImage),
	}

	go func() {

		for {
			time.Sleep(30 * time.Second)
			now := time.Now()
			for k, img := range s.polygonCache {
				if now.Sub(img.ts) > time.Minute*1 {
					delete(s.polygonCache, k)
				}
			}
		}
	}()

	return s
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

func (s *Service) GetPolygonMask(_ context.Context, groupId int, imageUid uuid.UUID) (image.Image, error) {
	const op = "service.GetPolygonMask"

	if cached, ok := s.polygonCache[imageUid]; ok {
		cached.ts = time.Now()
		return cached.img, nil
	}

	f, err := s.images.Open(groupId, imageUid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	p, err := s.polygons.DetectPolygons(img)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	mask := image.NewRGBA(img.Bounds())
	drawPolygon(mask, p)

	s.polygonCache[imageUid] = cachedImage{img: img, ts: time.Now()}
	return mask, nil
}
