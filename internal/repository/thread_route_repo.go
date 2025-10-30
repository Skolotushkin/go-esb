package repository

import (
	"context"

	"go-esb/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ThreadRouteRepository interface {
	GetThreadRoutes(ctx context.Context, threadID uuid.UUID) ([]models.ThreadRoute, error)
	GetThreadRouteByDirection(ctx context.Context, threadID uuid.UUID, direction models.Directions) ([]models.ThreadRoute, error)
	GetThreadRouteByRouteID(ctx context.Context, routeID uuid.UUID) (*models.ThreadRoute, error)
	CreateThreadRoute(ctx context.Context, tr *models.ThreadRoute) error
	GetThreadWithGroup(ctx context.Context, threadID uuid.UUID) (*models.Thread, *models.ThreadGroup, error)
}

type threadRouteRepository struct {
	db *sqlx.DB
}

func NewThreadRouteRepository(db *sqlx.DB) ThreadRouteRepository {
	return &threadRouteRepository{db: db}
}

func (r *threadRouteRepository) GetThreadRoutes(ctx context.Context, threadID uuid.UUID) ([]models.ThreadRoute, error) {
	var routes []models.ThreadRoute
	err := r.db.SelectContext(ctx, &routes, `
        SELECT thread, direction, route, file_format, object, routine 
        FROM thread_routes 
        WHERE thread = $1
    `, threadID)
	return routes, err
}

func (r *threadRouteRepository) GetThreadRouteByDirection(ctx context.Context, threadID uuid.UUID, direction models.Directions) ([]models.ThreadRoute, error) {
	var routes []models.ThreadRoute
	err := r.db.SelectContext(ctx, &routes, `
        SELECT thread, direction, route, file_format, object, routine 
        FROM thread_routes 
        WHERE thread = $1 AND direction = $2
    `, threadID, direction)
	return routes, err
}

func (r *threadRouteRepository) CreateThreadRoute(ctx context.Context, tr *models.ThreadRoute) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO thread_routes (thread, direction, route, file_format, object, routine)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (thread, direction, route) DO NOTHING
    `, tr.Thread, tr.Direction, tr.Route, tr.FileFormat, tr.Object, tr.Routine)
	return err
}

func (r *threadRouteRepository) GetThreadWithGroup(ctx context.Context, threadID uuid.UUID) (*models.Thread, *models.ThreadGroup, error) {
	var thread models.Thread
	err := r.db.GetContext(ctx, &thread, `
        SELECT ref, name, "group", message_convert_type 
        FROM threads 
        WHERE ref = $1
    `, threadID)
	if err != nil {
		return nil, nil, err
	}

	var group models.ThreadGroup
	err = r.db.GetContext(ctx, &group, `
        SELECT ref, name, protocol, parent, message_broker 
        FROM threads_groups 
        WHERE ref = $1
    `, thread.Group)
	if err != nil {
		return nil, nil, err
	}

	return &thread, &group, nil
}

