package repository

import (
	"context"

	"go-esb/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ThreadRepository interface {
	Create(ctx context.Context, thread *models.Thread) error
	GetAll(ctx context.Context) ([]models.Thread, error)
	GetByGroup(ctx context.Context, groupID uuid.UUID) ([]models.Thread, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type threadRepository struct {
	db *sqlx.DB
}

func NewThreadRepository(db *sqlx.DB) ThreadRepository {
	return &threadRepository{db: db}
}

func (r *threadRepository) Create(ctx context.Context, t *models.Thread) error {
	t.Ref = uuid.New()
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO threads (ref, name, "group", message_convert_type)
        VALUES ($1, $2, $3, $4)
    `, t.Ref, t.Name, t.Group, t.MessageConvertType)
	return err
}

func (r *threadRepository) GetAll(ctx context.Context) ([]models.Thread, error) {
	var threads []models.Thread
	err := r.db.SelectContext(ctx, &threads, `
        SELECT ref, name, "group", message_convert_type FROM threads ORDER BY name
    `)
	return threads, err
}

func (r *threadRepository) GetByGroup(ctx context.Context, groupID uuid.UUID) ([]models.Thread, error) {
	var threads []models.Thread
	err := r.db.SelectContext(ctx, &threads, `
        SELECT ref, name, "group", message_convert_type FROM threads WHERE "group" = $1
    `, groupID)
	return threads, err
}

func (r *threadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM threads WHERE ref = $1`, id)
	return err
}
