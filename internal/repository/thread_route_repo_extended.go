package repository

import (
	"context"

	"go-esb/internal/models"

	"github.com/google/uuid"
)

// GetThreadRouteByRouteID находит thread route по route ID и направлению
func (r *threadRouteRepository) GetThreadRouteByRouteID(ctx context.Context, routeID uuid.UUID) (*models.ThreadRoute, error) {
	var route models.ThreadRoute
	err := r.db.GetContext(ctx, &route, `
        SELECT thread, direction, route, file_format, object, routine 
        FROM thread_routes 
        WHERE route = $1
        LIMIT 1
    `, routeID)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

