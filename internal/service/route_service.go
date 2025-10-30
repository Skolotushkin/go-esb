package service

import (
	"context"
	"errors"

	"go-esb/internal/models"
	"go-esb/internal/repository"
)

type RouteService interface {
	Create(ctx context.Context, name, path string, method models.RestMethod, systemID string) (*models.Route, error)
	GetAll(ctx context.Context) ([]models.Route, error)
	GetBySystem(ctx context.Context, systemID string) ([]models.Route, error)
	Delete(ctx context.Context, id string) error
}

type routeService struct {
	repo    repository.RouteRepository
	sysRepo repository.SystemRepository
}

func NewRouteService(routeRepo repository.RouteRepository, systemRepo repository.SystemRepository) RouteService {
	return &routeService{repo: routeRepo, sysRepo: systemRepo}
}

func (s *routeService) Create(ctx context.Context, name, path string, method models.RestMethod, systemID string) (*models.Route, error) {
	if name == "" || path == "" {
		return nil, errors.New("name and path cannot be empty")
	}

	sysID, err := parseUUID(systemID)
	if err != nil {
		return nil, err
	}

	// Ensure the system exists
	_, err = s.sysRepo.GetByID(ctx, sysID)
	if err != nil {
		return nil, errors.New("system not found")
	}

	route := &models.Route{
		Name:   name,
		Path:   path,
		Method: method,
		System: sysID,
	}

	if err := s.repo.Create(ctx, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *routeService) GetAll(ctx context.Context) ([]models.Route, error) {
	return s.repo.GetAll(ctx)
}

func (s *routeService) GetBySystem(ctx context.Context, systemID string) ([]models.Route, error) {
	sysID, err := parseUUID(systemID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetBySystem(ctx, sysID)
}

func (s *routeService) Delete(ctx context.Context, id string) error {
	routeID, err := parseUUID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, routeID)
}
