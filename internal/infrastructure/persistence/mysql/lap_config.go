package mysql

import (
	"FairLAP/internal/domain/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

type LapConfigRepo struct {
	db *sqlx.DB
}

func NewLapConfigRepo(db *sqlx.DB) *LapConfigRepo {
	return &LapConfigRepo{
		db: db,
	}
}

func (r *LapConfigRepo) SaveParameter(ctx context.Context, lapId string, params entity.LapParameter) error {
	const op = "LapConfigRepo.AddParameter"

	query := `
INSERT INTO lap_config (lap_id, class, value) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE value=?`

	if _, err := r.db.ExecContext(ctx, query, lapId, params.Class, params.Value, params.Value); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *LapConfigRepo) DeleteParameters(ctx context.Context, lapId string, classes []string) error {
	const op = "LapConfigRepo.DeleteParameters"

	placeHolders := strings.TrimSuffix(strings.Repeat("?,", len(classes)), ",")

	query := "DELETE FROM lap_config WHERE lap_id=? AND class IN (" + placeHolders + ")"

	args := make([]any, len(classes)+1)
	args[0] = lapId

	for i, class := range classes {
		args[i+1] = class
	}

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *LapConfigRepo) UpdateParameter(ctx context.Context, lapId string, param entity.LapParameter) error {
	const op = "LapConfigRepo.UpdateParameters"
	if _, err := r.db.ExecContext(ctx, "UPDATE lap_config SET value=? WHERE lap_id=? AND class=?", param.Value, lapId, param.Class); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *LapConfigRepo) GetLapParameters(ctx context.Context, lapId string) ([]entity.LapParameter, error) {
	const op = "LapConfigRepo.GetConfig"

	var params []entity.LapParameter
	if err := r.db.SelectContext(ctx, &params, "SELECT class, value FROM lap_config WHERE lap_id=?", lapId); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return params, nil
}
