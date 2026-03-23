package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/tasks"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/hibiken/asynq"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type TaskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error)
	Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	txManager  txmanager.TxManager
	taskClient TaskEnqueuer
	logger     *zap.SugaredLogger
}

func NewUserService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	txManager txmanager.TxManager,
	taskClient TaskEnqueuer,
	logger *zap.SugaredLogger,
) UserService {
	return &userService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		txManager:  txManager,
		taskClient: taskClient,
		logger:     logger,
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

func (s *userService) List(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.List(ctx)
}

func (s *userService) Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("userService.Create hash: %w", err)
	}

	now := time.Now()
	tenant := &domain.Tenant{
		ID:        ksuid.New().String(),
		Name:      req.Name,
		Slug:      domain.NewSlug(req.Name),
		Phone:     req.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	user := &domain.User{
		ID:        ksuid.New().String(),
		Email:     req.Email,
		Password:  string(hashed),
		Name:      req.Name,
		Phone:     req.Phone,
		TenantID:  tenant.ID,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := s.tenantRepo.Create(ctx, tenant); err != nil {
			return err
		}
		return s.userRepo.Create(ctx, user)
	}); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("userService.Create: %w", err)
	}

	if s.taskClient != nil {
		task, err := tasks.NewWelcomeEmailTask(tasks.WelcomeEmailPayload{
			UserID: user.ID,
			Email:  user.Email,
			Name:   user.Name,
		})
		if err != nil {
			s.logger.Warnw("failed to create welcome email task", "error", err, "userID", user.ID)
		} else {
			if _, err := s.taskClient.Enqueue(task); err != nil {
				s.logger.Warnw("failed to enqueue welcome email", "error", err, "userID", user.ID)
			}
		}
	}

	return user, nil
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
