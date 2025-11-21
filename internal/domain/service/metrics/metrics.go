package metrics

import (
	"FairLAP/internal/domain/aggregate"
	"FairLAP/internal/domain/entity"
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type DetectionsRepo interface {
	GetByGroup(ctx context.Context, group int) ([]entity.Detection, error)
}

type GroupsRepo interface {
	GetLaps(ctx context.Context) ([]aggregate.LapLastDetect, error)
	GetLapId(ctx context.Context, groupId int) (string, error)
}

type ConfigService interface {
	GetConfig(ctx context.Context, lapId string) (map[string]int, error)
}

type Service struct {
	groups     GroupsRepo
	detections DetectionsRepo
	lapConfig  ConfigService
}

func NewService(groups GroupsRepo, detections DetectionsRepo, lapConfig ConfigService) *Service {
	return &Service{
		groups:     groups,
		detections: detections,
		lapConfig:  lapConfig,
	}
}

type LapItem struct {
	HaveProblems bool      `json:"have_problems"`
	LastGroup    int       `json:"last_group"`
	LastDetect   time.Time `json:"last_detect"`
}

func (s *Service) GetLaps(ctx context.Context) (map[string]LapItem, error) {
	const op = "metrics_service.GetLaps"

	laps, err := s.groups.GetLaps(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	lapMap := make(map[string]LapItem)

	for _, lap := range laps {
		lapItem := LapItem{
			LastGroup:  lap.LastGroup,
			LastDetect: lap.LastDetect,
		}

		config, err := s.lapConfig.GetConfig(ctx, lap.LapId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		detections, err := s.detections.GetByGroup(ctx, lap.LastGroup)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		maxSum := config["sum"]
		deletionsLevelSum := 0

		for _, detection := range detections {
			deletionsLevelSum += config[detection.Class]
			if deletionsLevelSum >= maxSum {
				lapItem.HaveProblems = true
				break
			}
		}

		lapMap[lap.LapId] = lapItem
	}

	return lapMap, nil
}

type GroupMetric struct {
	ImageCount      int                           `json:"image_count"`
	DetectionsCount int                           `json:"detections_count"`
	Classes         []string                      `json:"classes"`
	Detections      map[string][]entity.Detection `json:"detections"`
}

func (s *Service) GetGroupMetric(ctx context.Context, groupId int) (*GroupMetric, error) {
	const op = "metrics_service.GetGroupMetric"

	detections, err := s.detections.GetByGroup(ctx, groupId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	imagesMap := make(map[uuid.UUID]struct{})

	metric := &GroupMetric{
		Detections: make(map[string][]entity.Detection),
	}

	for _, detection := range detections {
		imagesMap[detection.ImageUid] = struct{}{}
		metric.Detections[detection.Class] = append(metric.Detections[detection.Class], detection)
	}

	metric.ImageCount = len(imagesMap)

	metric.Classes = make([]string, 0, len(metric.Detections))

	for class := range metric.Detections {
		metric.Classes = append(metric.Classes, class)
	}

	return metric, nil
}

type GroupMetricV2 struct {
	ImageCount      int                            `json:"image_count"`
	DetectionsCount int                            `json:"detections_count"`
	Images          map[uuid.UUID][]ImageDetection `json:"images"`
}

type ImageDetection struct {
	Id          int    `json:"id" db:"id"`
	Class       string `json:"class" db:"class"`
	DamageLevel int    `json:"damage_level" db:"damage_level"`
}

func (s *Service) GetGroupMetricV2(ctx context.Context, groupId int) (*GroupMetricV2, error) {
	const op = "metrics_service.GetGroupMetric"

	lapId, err := s.groups.GetLapId(ctx, groupId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	config, err := s.lapConfig.GetConfig(ctx, lapId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	detections, err := s.detections.GetByGroup(ctx, groupId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	metric := &GroupMetricV2{
		DetectionsCount: len(detections),
		Images:          make(map[uuid.UUID][]ImageDetection),
	}

	for _, detection := range detections {

		metric.Images[detection.ImageUid] = append(metric.Images[detection.ImageUid], ImageDetection{
			Id:          detection.Id,
			Class:       detection.Class,
			DamageLevel: config[detection.Class],
		})
	}

	metric.ImageCount = len(metric.Images)

	return metric, nil
}
