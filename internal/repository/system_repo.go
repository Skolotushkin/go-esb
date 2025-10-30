package repository

import (
	"context"

	"go-esb/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SystemRepository interface {
	Create(ctx context.Context, system *models.System) error
	GetAll(ctx context.Context) ([]models.System, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.System, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, system *models.System) error
}

type systemRepository struct {
	db *sqlx.DB
}

func NewSystemRepository(db *sqlx.DB) SystemRepository {
	return &systemRepository{db: db}
}

func (r *systemRepository) Create(ctx context.Context, s *models.System) error {
	s.Ref = uuid.New()
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO systems (ref, name)
        VALUES ($1, $2)
    `, s.Ref, s.Name)
	return err
}

func (r *systemRepository) GetAll(ctx context.Context) ([]models.System, error) {
	var systems []models.System
	err := r.db.SelectContext(ctx, &systems, `SELECT ref, name FROM systems ORDER BY name`)
	return systems, err
}

func (r *systemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.System, error) {
	var system models.System
	err := r.db.GetContext(ctx, &system, `SELECT ref, name FROM systems WHERE ref = $1`, id)
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (r *systemRepository) Update(ctx context.Context, s *models.System) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE systems SET name = $2 WHERE ref = $1
    `, s.Ref, s.Name)
	return err
}

func (r *systemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM systems WHERE ref = $1`, id)
	return err
}
