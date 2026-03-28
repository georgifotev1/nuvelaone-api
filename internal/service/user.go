package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/hibiken/asynq"
)

type TaskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context, tenandID string) ([]domain.User, error)
	Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(
	userRepo repository.UserRepository,
) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("userService.GetByID: %w", err)
	}
	return user, nil
}

func (s *userService) List(ctx context.Context, tenandID string) ([]domain.User, error) {
	return s.userRepo.List(ctx, tenandID)
}

func (s *userService) Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("userService.Update: %w", err)
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("userService.Update: %w", err)
	}
	return user, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("userService.Delete: %w", err)
	}
	return nil
}
