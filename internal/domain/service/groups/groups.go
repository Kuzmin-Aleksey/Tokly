package groups

import (
	"FairLAP/internal/domain/entity"
	"FairLAP/pkg/contextx"
	"FairLAP/pkg/logx"
	"context"
	"fmt"
	"time"
)

type Repo interface {
	Save(ctx context.Context, group *entity.Group) error
	Delete(ctx context.Context, id int) error
}

type ImagesDeleter interface {
	DeleteGroup(groupId int) error
}

type Service struct {
	repo   Repo
	images ImagesDeleter
}

func NewService(repo Repo, images ImagesDeleter) *Service {
	return &Service{
		repo:   repo,
		images: images,
	}
}

func (s *Service) CreateGroup(ctx context.Context, lapId int) (int, error) {
	const op = "groups_service.CreateGroup"

	group := &entity.Group{
		LapId:    lapId,
		CreateAt: time.Now().In(time.UTC),
	}

	if err := s.repo.Save(ctx, group); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return group.Id, nil
}

func (s *Service) DeleteGroup(ctx context.Context, id int) error {
	const op = "groups_service.DeleteGroup"
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.images.DeleteGroup(id); err != nil {
		contextx.GetLoggerOrDefault(ctx).ErrorContext(ctx, "delete group image", logx.Error(err))
	}

	return nil
}
