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

type DetectionsRepo struct {
	db *sqlx.DB
}

func NewDetectionsRepo(db *sqlx.DB) *DetectionsRepo {
	return &DetectionsRepo{
		db: db,
	}
}

func (r *DetectionsRepo) Save(ctx context.Context, detections *entity.Detection) error {
	const op = "DetectionsRepo.Save"
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO detections (group_id, image_uid, class) VALUES (:group_id, :image_uid, :class)", detections)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	detections.Id = int(id)

	return nil
}

func (r *DetectionsRepo) GetByGroup(ctx context.Context, group int) ([]entity.Detection, error) {
	const op = "DetectionsRepo.GetByGroup"
	var detections []entity.Detection

	if err := r.db.SelectContext(ctx, &detections, "SELECT * FROM detections WHERE group_id=?", group); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return detections, nil
}

func (r *DetectionsRepo) IsExistProblem(ctx context.Context, lapId int) (bool, error) {
	const op = "DetectionsRepo.IsExistProblem"

	query := "SELECT EXISTS(SELECT * FROM detections INNER JOIN `groups` ON detections.group_id = `groups`.id WHERE detections.is_problem AND `groups`.lap_id=?)"
	var exist bool
	if err := r.db.QueryRowContext(ctx, query, lapId).Scan(&exist); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exist, nil
}

func (r *DetectionsRepo) SaveRects(ctx context.Context, rects []entity.RectDetection) error {
	const op = "DetectionsRepo.Save"
	query := "INSERT INTO detection_rects (detection_id, width, height, x0, y0, x1, y1, confidence) VALUES"
	args := make([]any, 0, len(rects)*8)

	for _, rect := range rects {
		query += " (?, ?, ?, ?, ?, ?, ?, ?),"
		args = append(args, rect.DetectionId, rect.Width, rect.Height, rect.X0, rect.Y0, rect.X1, rect.Y1, rect.Confidence)
	}

	query = strings.TrimSuffix(query, ",")

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *DetectionsRepo) GetRect(ctx context.Context, detectionId int) (*entity.RectDetection, string, error) {
	const op = "DetectionsRepo.GetRect"
	var rect entity.RectDetection
	var class string
	if err := r.db.QueryRowContext(ctx, "SELECT detection_rects.id, detection_rects.detection_id, detection_rects.width, detection_rects.height, detection_rects.x0, detection_rects.y0, detection_rects.x1, detection_rects.y1, detection_rects.confidence, detections.class FROM detection_rects INNER JOIN detections ON detection_rects.detection_id = detections.id WHERE detection_id=?", detectionId).Scan(
		&rect.Id, &rect.DetectionId, &rect.Width, &rect.Height, &rect.X0, &rect.Y0, &rect.X1, &rect.Y1, &rect.Confidence, &class); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("%s: %w", op, err)
		}
	}

	return &rect, class, nil
}
