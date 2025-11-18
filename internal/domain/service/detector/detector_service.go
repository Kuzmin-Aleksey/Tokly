package detector

import (
	"FairLAP/internal/config"
	"FairLAP/internal/domain/entity"
	"FairLAP/pkg/yolo_model"
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"slices"
)

type Repo interface {
	Save(ctx context.Context, detections []entity.Detection) error
}

type ImageRepo interface {
	Save(groupId int, img image.Image) (uuid.UUID, error)
}

type Service struct {
	problemClasses map[int]struct{}
	model          *yolo_model.Model
	repo           Repo
	images         ImageRepo
}

func NewService(model *yolo_model.Model, cfg *config.DetectorConfig, repo Repo, images ImageRepo) *Service {
	problemClasses := make(map[int]struct{})

	for _, classId := range cfg.ProblemClasses {
		problemClasses[classId] = struct{}{}
	}

	return &Service{
		model:          model,
		repo:           repo,
		images:         images,
		problemClasses: problemClasses,
	}
}

func (s *Service) Detect(ctx context.Context, groupId int, img image.Image) error {
	const op = "detector_service.Detect"

	modelsDetections, err := s.model.Detect(img)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	renderedImg := s.drawDetected(img, modelsDetections)

	imgUid, err := s.images.Save(groupId, renderedImg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	detections := make([]entity.Detection, len(modelsDetections))

	for i, detection := range modelsDetections {
		_, isProblem := s.problemClasses[detection.ClassID]

		detections[i] = entity.Detection{
			GroupId:   groupId,
			ImageUid:  imgUid,
			Class:     detection.ClassName,
			IsProblem: isProblem,
		}
	}

	if err := s.repo.Save(ctx, detections); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

var opentypeFont, _ = opentype.Parse(fontTTF)

func (s *Service) drawDetected(img image.Image, detections []yolo_model.Detection) image.Image {
	res := image.NewRGBA(img.Bounds())
	draw.Draw(res, img.Bounds(), img, image.Point{}, draw.Src)

	fontSize := float64(img.Bounds().Dy()) / 50

	face, _ := opentype.NewFace(opentypeFont, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	fontHeight := face.Metrics().Height.Floor()

	drawer := &font.Drawer{
		Dst:  res,
		Src:  image.NewUniform(color.White),
		Face: face,
	}

	fontRects := make([]image.Rectangle, 0, len(detections))

	slices.SortFunc(detections, func(d1, d2 yolo_model.Detection) int {
		return d1.BBox.Min.Y - d2.BBox.Min.Y
	})

	thickness := int(float64(img.Bounds().Dy()) / 900)

	for _, detection := range detections {
		var c color.Color
		if _, isProblem := s.problemClasses[detection.ClassID]; isProblem {
			c = color.RGBA{255, 0, 0, 0}
		} else {
			c = color.RGBA{0, 0, 255, 0}
		}

		drawRect(res, c, detection.BBox, thickness)

		lbl := fmt.Sprintf("%s: %0.2f", detection.ClassName, detection.Confidence)

		width := font.MeasureString(face, lbl).Ceil()

		x, y := detection.BBox.Min.X+thickness+2, detection.BBox.Min.Y+fontHeight+thickness

		fontRect := image.Rect(x, y-fontHeight, x+width, y)

		for _, rect := range fontRects {
			if rect.Overlaps(fontRect) {
				fontRect.Min.Y += fontHeight
				fontRect.Max.Y += fontHeight
			}
		}
		drawer.Dot = fixed.P(fontRect.Min.X, fontRect.Max.Y)
		drawer.DrawString(lbl)

		fontRects = append(fontRects, fontRect)
	}

	return res
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
