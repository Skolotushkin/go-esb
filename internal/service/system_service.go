package service

import (
	"context"
	"errors"
	"go-esb/internal/models"
	"go-esb/internal/repository"
)

type SystemService interface {
	Create(ctx context.Context, name string) (*models.System, error)
	GetAll(ctx context.Context) ([]models.System, error)
	GetByID(ctx context.Context, id string) (*models.System, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, name string) error
}

type systemService struct {
	repo repository.SystemRepository
}

func NewSystemService(repo repository.SystemRepository) SystemService {
	return &systemService{repo: repo}
}

func (s *systemService) Create(ctx context.Context, name string) (*models.System, error) {
	if name == "" {
		return nil, errors.New("system name cannot be empty")
	}

	sys := &models.System{Name: name}
	if err := s.repo.Create(ctx, sys); err != nil {
		return nil, err
	}
	return sys, nil
}

func (s *systemService) GetAll(ctx context.Context) ([]models.System, error) {
	return s.repo.GetAll(ctx)
}

func (s *systemService) GetByID(ctx context.Context, id string) (*models.System, error) {
	sysID, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, sysID)
}

func (s *systemService) Delete(ctx context.Context, id string) error {
	sysID, err := parseUUID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, sysID)
}

func (s *systemService) Update(ctx context.Context, id string, name string) error {
	sysID, err := parseUUID(id)
	if err != nil {
		return err
	}
	sys := &models.System{Ref: sysID, Name: name}
	return s.repo.Update(ctx, sys)
}
