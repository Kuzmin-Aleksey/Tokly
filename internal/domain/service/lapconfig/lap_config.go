package lapconfig

import (
	"FairLAP/internal/domain/entity"
	"context"
	"fmt"
	"maps"
)

type Repo interface {
	SaveParameter(ctx context.Context, lapId string, params entity.LapParameter) error
	GetLapParameters(ctx context.Context, lapId string) ([]entity.LapParameter, error)
	DeleteParameters(ctx context.Context, lapId string, classes []string) error
}

type Service struct {
	repo          Repo
	defaultConfig map[string]int
}

func NewService(repo Repo, defaultConfig map[string]int) *Service {
	return &Service{
		repo:          repo,
		defaultConfig: defaultConfig,
	}
}

func (s *Service) SaveDefaultConfig(ctx context.Context, lapId string) error {
	const op = "lap_config.SaveDefaultConfig"

	for class, val := range s.defaultConfig {
		param := entity.LapParameter{
			Class: class,
			Value: val,
		}

		if err := s.repo.SaveParameter(ctx, lapId, param); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Service) SaveLapConfig(ctx context.Context, lapId string, params map[string]int) error {
	const op = "lap_config.SaveLapConfig"

	currentParams, isDefault, err := s.getConfig(ctx, lapId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if isDefault {
		if err := s.SaveDefaultConfig(ctx, lapId); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	var forDelete []string

	for param := range params {
		delete(currentParams, param)

		param := entity.LapParameter{
			Class: param,
			Value: params[param],
		}

		if err := s.repo.SaveParameter(ctx, lapId, param); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	for param := range currentParams {
		forDelete = append(forDelete, param)
	}

	if len(forDelete) > 0 {
		if err := s.repo.DeleteParameters(ctx, lapId, forDelete); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func (s *Service) getConfig(ctx context.Context, lapId string) (map[string]int, bool, error) {
	const op = "lap_config.getConfig"
	lapParameters, err := s.repo.GetLapParameters(ctx, lapId)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	mapParams := make(map[string]int)
	for _, param := range lapParameters {
		mapParams[param.Class] = param.Value
	}

	if len(lapParameters) == 0 {
		maps.Copy(mapParams, s.defaultConfig)
		return mapParams, true, nil
	}

	return mapParams, false, nil
}

func (s *Service) GetConfig(ctx context.Context, lapId string) (map[string]int, error) {
	const op = "lap_config.GetConfig"
	cfg, _, err := s.getConfig(ctx, lapId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return cfg, nil
}
