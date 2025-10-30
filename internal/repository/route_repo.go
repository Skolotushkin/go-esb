package repository

import (
	"context"

	"go-esb/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RouteRepository interface {
	Create(ctx context.Context, route *models.Route) error
	GetAll(ctx context.Context) ([]models.Route, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Route, error)
	GetBySystem(ctx context.Context, systemID uuid.UUID) ([]models.Route, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type routeRepository struct {
	db *sqlx.DB
}

func NewRouteRepository(db *sqlx.DB) RouteRepository {
	return &routeRepository{db: db}
}

func (r *routeRepository) Create(ctx context.Context, route *models.Route) error {
	route.Ref = uuid.New()
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO routes (ref, name, path, system, method)
        VALUES ($1, $2, $3, $4, $5)
    `, route.Ref, route.Name, route.Path, route.System, route.Method)
	return err
}

func (r *routeRepository) GetAll(ctx context.Context) ([]models.Route, error) {
	var routes []models.Route
	err := r.db.SelectContext(ctx, &routes, `
        SELECT ref, name, path, system, method FROM routes ORDER BY name
    `)
	return routes, err
}

func (r *routeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Route, error) {
	var route models.Route
	err := r.db.GetContext(ctx, &route, `
        SELECT ref, name, path, system, method FROM routes WHERE ref = $1
    `, id)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *routeRepository) GetBySystem(ctx context.Context, systemID uuid.UUID) ([]models.Route, error) {
	var routes []models.Route
	err := r.db.SelectContext(ctx, &routes, `
        SELECT ref, name, path, system, method FROM routes WHERE system = $1
    `, systemID)
	return routes, err
}

func (r *routeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM routes WHERE ref = $1`, id)
	return err
}
