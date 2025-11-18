package mysql

import (
	"FairLAP/internal/domain/entity"
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

func (r *GroupsRepo) GetAll(ctx context.Context) ([]entity.Group, error) {
	const op = "DetectionsRepo.GetAll"
	var groups []entity.Group
	if err := r.db.SelectContext(ctx, &groups, "SELECT * FROM `groups`"); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	return groups, nil
}

func (r *GroupsRepo) Delete(ctx context.Context, id int) error {
	const op = "DetectionsRepo.DeleteGroup"
	if _, err := r.db.ExecContext(ctx, "DELETE FROM `groups` WHERE id = ?", id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
