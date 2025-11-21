package mysql

import (
	"FairLAP/internal/domain/aggregate"
	"FairLAP/internal/domain/entity"
	"FairLAP/pkg/failure"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type GroupsRepo struct {
	db *sqlx.DB
}

func NewGroupsRepo(db *sqlx.DB) *GroupsRepo {
	return &GroupsRepo{
		db: db,
	}
}

func (r *GroupsRepo) Save(ctx context.Context, group *entity.Group) error {
	const op = "DetectionsRepo.SaveGroup"

	res, err := r.db.NamedExecContext(ctx, "INSERT INTO `groups` (lap_id, create_at) VALUES (:lap_id, :create_at)", group)
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}

	group.Id = int(id)

	return nil
}

func (r *GroupsRepo) GetByLap(ctx context.Context, lapId string) ([]entity.Group, error) {
	const op = "DetectionsRepo.GetByLap"
	var groups []entity.Group
	if err := r.db.SelectContext(ctx, &groups, "SELECT * FROM `groups` WHERE lap_id=?", lapId); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	return groups, nil
}

func (r *GroupsRepo) GetLaps(ctx context.Context) ([]aggregate.LapLastDetect, error) {
	const op = "DetectionsRepo.GetLaps"
	var laps []aggregate.LapLastDetect
	if err := r.db.SelectContext(ctx, &laps, "SELECT max(id) AS last_group, lap_id, max(create_at) AS last_detect FROM `groups` GROUP BY lap_id"); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	return laps, nil
}

func (r *GroupsRepo) GetLapId(ctx context.Context, groupId int) (string, error) {
	const op = "DetectionsRepo.GetLaps"
	var laps string
	if err := r.db.GetContext(ctx, &laps, "SELECT lap_id FROM `groups` WHERE id=?", groupId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, failure.NewNotFoundError(err.Error()))
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return laps, nil
}

func (r *GroupsRepo) Delete(ctx context.Context, id int) error {
	const op = "DetectionsRepo.DeleteGroup"
	if _, err := r.db.ExecContext(ctx, "DELETE FROM `groups` WHERE id = ?", id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
