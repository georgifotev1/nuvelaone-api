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
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService interface {
	Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error)
}

type authService struct {
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	tokenRepo  repository.TokenRepository
	txManager  txmanager.TxManager
	taskClient TaskEnqueuer
	logger     *zap.SugaredLogger
	cfg        AuthConfig
}

type AuthConfig struct {
	AccessSecret    string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	tokenRepo repository.TokenRepository,
	txManager txmanager.TxManager,
	taskClient TaskEnqueuer,
	logger *zap.SugaredLogger,
	cfg AuthConfig,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		tokenRepo:  tokenRepo,
		txManager:  txManager,
		taskClient: taskClient,
		logger:     logger,
		cfg:        cfg,
	}
}

func (s *authService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("authService.Register hash: %w", err)
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
		Role:      domain.RoleOwner,
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
		return nil, fmt.Errorf("authService.Register tx: %w", err)
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

func (s *authService) Login(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, user)
}

func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := auth.HashToken(rawRefreshToken)
	if err := s.tokenRepo.Revoke(ctx, hash); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidToken
		}
		return fmt.Errorf("authService.Logout: %w", err)
	}
	return nil
}

func (s *authService) Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error) {
	hash := auth.HashToken(rawRefreshToken)

	stored, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if stored.RevokedAt != nil {
		// possible token theft — revoke all tokens for this user
		s.logger.Warnw("reuse of revoked refresh token", "userID", stored.UserID)
		if err := s.tokenRepo.RevokeAllForUser(ctx, stored.UserID); err != nil {
			s.logger.Errorw("failed to revoke all tokens", "error", err, "userID", stored.UserID)
		}
		return nil, ErrInvalidToken
	}
	if stored.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var pair *domain.TokenPair
	if err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := s.tokenRepo.Revoke(ctx, hash); err != nil {
			return err
		}
		var err error
		pair, err = s.issueTokenPair(ctx, user)
		return err
	}); err != nil {
		return nil, fmt.Errorf("authService.Refresh: %w", err)
	}

	return pair, nil
}

func (s *authService) issueTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	accessToken, err := auth.GenerateAccessToken(user, s.cfg.AccessSecret, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("issueTokenPair access: %w", err)
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("issueTokenPair refresh: %w", err)
	}

	if err := s.tokenRepo.Store(ctx, &domain.RefreshToken{
		ID:        ksuid.New().String(),
		UserID:    user.ID,
		TokenHash: auth.HashToken(rawRefresh),
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("issueTokenPair store: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}
