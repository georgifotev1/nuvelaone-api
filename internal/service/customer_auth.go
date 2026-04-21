package service

import (
	"context"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type CustomerAuthService interface {
	Register(ctx context.Context, tenantID string, req domain.CustomerRegisterRequest) (*domain.Customer, error)
	Login(ctx context.Context, tenantID string, req domain.CustomerLoginRequest) (*domain.TokenPair, error)
	Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, rawRefreshToken string) error
}

type CustomerAuthConfig struct {
	AccessSecret    string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type customerAuthService struct {
	customerRepo repository.CustomerRepository
	tokenRepo    repository.TokenRepository
	logger       *zap.SugaredLogger
	cfg          CustomerAuthConfig
}

func NewCustomerAuthService(
	customerRepo repository.CustomerRepository,
	tokenRepo repository.TokenRepository,
	logger *zap.SugaredLogger,
	cfg CustomerAuthConfig,
) CustomerAuthService {
	return &customerAuthService{
		customerRepo: customerRepo,
		tokenRepo:    tokenRepo,
		logger:       logger,
		cfg:          cfg,
	}
}

func (s *customerAuthService) Register(ctx context.Context, tenantID string, req domain.CustomerRegisterRequest) (*domain.Customer, error) {
	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Register hash: %w", err)
	}

	now := time.Now()
	hashedStr := string(hashed)
	customer := &domain.Customer{
		ID:        ksuid.New().String(),
		TenantID:  tenantID,
		Name:      req.Name,
		Email:     &req.Email,
		Password:  &hashedStr,
		Phone:     req.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("customerAuthService.Register: %w", err)
	}

	return customer, nil
}

func (s *customerAuthService) Login(ctx context.Context, tenantID string, req domain.CustomerLoginRequest) (*domain.TokenPair, error) {
	customer, err := s.customerRepo.GetByEmail(ctx, tenantID, req.Email)
	if err != nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	if customer.Password == nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	if !auth.CheckPassword(req.Password, *customer.Password) {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	return s.issueTokenPair(ctx, customer)
}

func (s *customerAuthService) Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error) {
	hash := auth.HashToken(rawRefreshToken)

	stored, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, apperr.Unauthorized("invalid token")
	}
	if stored.RevokedAt != nil {
		s.logger.Warnw("reuse of revoked refresh token", "entityID", stored.EntityID)
		if stored.EntityType == domain.TokenEntityCustomer {
			if err := s.tokenRepo.RevokeAllForCustomer(ctx, stored.EntityID); err != nil {
				s.logger.Errorw("failed to revoke all tokens", "error", err, "entityID", stored.EntityID)
			}
		}
		return nil, apperr.Unauthorized("invalid token")
	}
	if stored.ExpiresAt.Before(time.Now()) {
		return nil, apperr.Unauthorized("token expired")
	}

	if stored.EntityType != domain.TokenEntityCustomer {
		return nil, apperr.Unauthorized("invalid token")
	}

	customer, err := s.customerRepo.GetByID(ctx, stored.TenantID, stored.EntityID)
	if err != nil {
		return nil, apperr.Unauthorized("invalid token")
	}

	var pair *domain.TokenPair
	if err := s.tokenRepo.Revoke(ctx, hash); err != nil {
		return nil, err
	}
	pair, err = s.issueTokenPair(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Refresh: %w", err)
	}

	return pair, nil
}

func (s *customerAuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := auth.HashToken(rawRefreshToken)
	if err := s.tokenRepo.Revoke(ctx, hash); err != nil {
		return fmt.Errorf("customerAuthService.Logout: %w", apperr.Unauthorized("invalid token"))
	}
	return nil
}

func (s *customerAuthService) issueTokenPair(ctx context.Context, customer *domain.Customer) (*domain.TokenPair, error) {
	accessToken, err := auth.GenerateCustomerAccessToken(customer, s.cfg.AccessSecret, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("issueTokenPair access: %w", err)
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("issueTokenPair refresh: %w", err)
	}

	if err := s.tokenRepo.Store(ctx, &domain.RefreshToken{
		ID:         ksuid.New().String(),
		EntityID:   customer.ID,
		EntityType: domain.TokenEntityCustomer,
		TenantID:   customer.TenantID,
		TokenHash:  auth.HashToken(rawRefresh),
		ExpiresAt:  time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt:  time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("issueTokenPair store: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}
