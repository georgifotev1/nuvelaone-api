package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/tasks"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/hibiken/asynq"
)

// ErrNotFound is returned when a resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a resource already exists.
var ErrConflict = errors.New("already exists")

// UserService defines the business operations for users.
type UserService interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error)
	Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	repo       repository.UserRepository
	taskClient *asynq.Client
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository, taskClient *asynq.Client) UserService {
	return &userService{repo: repo, taskClient: taskClient}
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("userService.GetByID: %w", ErrNotFound)
	}
	return user, nil
}

func (s *userService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *userService) Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	// Check for duplicate email
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrConflict
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("userService.Create hash: %w", err)
	}

	user := &domain.User{
		Email:    req.Email,
		Password: string(hashed),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("userService.Create: %w", err)
	}

	if s.taskClient != nil {
		task, err := tasks.NewWelcomeEmailTask(tasks.WelcomeEmailPayload{
			UserID: user.ID,
			Email:  user.Email,
			Name:   user.Name,
		})
		if err != nil {
			s.taskClient.Enqueue(task)
		}
	}

	return user, nil
}

func (s *userService) Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("userService.Update: %w", err)
	}
	return user, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("userService.Delete: %w", err)
	}
	return nil
}
