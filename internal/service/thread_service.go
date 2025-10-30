package service

import (
	"context"
	"errors"

	"go-esb/internal/models"
	"go-esb/internal/repository"
)

type ThreadService interface {
	Create(ctx context.Context, name string, groupID string, convType models.MessageConvertType) (*models.Thread, error)
	GetAll(ctx context.Context) ([]models.Thread, error)
	GetByGroup(ctx context.Context, groupID string) ([]models.Thread, error)
	Delete(ctx context.Context, id string) error
}

type threadService struct {
	repo repository.ThreadRepository
}

func NewThreadService(repo repository.ThreadRepository) ThreadService {
	return &threadService{repo: repo}
}

func (s *threadService) Create(ctx context.Context, name string, groupID string, convType models.MessageConvertType) (*models.Thread, error) {
	if name == "" {
		return nil, errors.New("thread name cannot be empty")
	}

	grpID, err := parseUUID(groupID)
	if err != nil {
		return nil, err
	}

	thread := &models.Thread{
		Name:               name,
		Group:              grpID,
		MessageConvertType: convType,
	}

	if err := s.repo.Create(ctx, thread); err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *threadService) GetAll(ctx context.Context) ([]models.Thread, error) {
	return s.repo.GetAll(ctx)
}

func (s *threadService) GetByGroup(ctx context.Context, groupID string) ([]models.Thread, error) {
	grpID, err := parseUUID(groupID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByGroup(ctx, grpID)
}

func (s *threadService) Delete(ctx context.Context, id string) error {
	threadID, err := parseUUID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, threadID)
}
