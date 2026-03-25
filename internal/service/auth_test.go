package service

import (
	"context"
	"testing"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/tasks"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()

	tests := []struct {
		name        string
		req         domain.RegisterRequest
		setupMocks  func(*repository.MockUserRepository, *repository.MockTenantRepository, *txmanager.MockTxManager, *tasks.MockTaskEnqueuer)
		expectedErr error
	}{
		{
			name: "success - creates tenant and user",
			req: domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
				Phone:    "+1234567890",
			},
			setupMocks: func(userRepo *repository.MockUserRepository, tenantRepo *repository.MockTenantRepository, txManager *txmanager.MockTxManager, taskClient *tasks.MockTaskEnqueuer) {
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(context.Background())
				})
				tenantRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				userRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				taskClient.On("Enqueue", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "conflict - duplicate email",
			req: domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			setupMocks: func(userRepo *repository.MockUserRepository, tenantRepo *repository.MockTenantRepository, txManager *txmanager.MockTxManager, taskClient *tasks.MockTaskEnqueuer) {
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(repository.ErrDuplicate).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(context.Background())
				})
				tenantRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				userRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(repository.MockUserRepository)
			tenantRepo := new(repository.MockTenantRepository)
			txManager := new(txmanager.MockTxManager)
			taskClient := new(tasks.MockTaskEnqueuer)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, tenantRepo, txManager, taskClient)
			}

			svc := NewAuthService(
				userRepo,
				tenantRepo,
				nil,
				txManager,
				taskClient,
				logger,
				AuthConfig{
					AccessSecret:    "test-secret",
					AccessTokenTTL:  time.Hour,
					RefreshTokenTTL: time.Hour * 24 * 7,
				},
			)

			user, err := svc.Register(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.Equal(t, tt.req.Name, user.Name)
			}

			userRepo.AssertExpectations(t)
			tenantRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
			taskClient.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()

	userID := ksuid.New().String()
	now := time.Now()
	userEmail := "test@example.com"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &domain.User{
		ID:        userID,
		Email:     userEmail,
		Password:  string(hashedPassword),
		Name:      "Test User",
		TenantID:  ksuid.New().String(),
		Role:      domain.RoleOwner,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name        string
		req         domain.LoginRequest
		setupMocks  func(*repository.MockUserRepository, *repository.MockTokenRepository)
		expectedErr error
	}{
		{
			name: "success",
			req: domain.LoginRequest{
				Email:    userEmail,
				Password: "password123",
			},
			setupMocks: func(userRepo *repository.MockUserRepository, tokenRepo *repository.MockTokenRepository) {
				userRepo.On("GetByEmail", context.Background(), userEmail).Return(testUser, nil)
				tokenRepo.On("Store", context.Background(), mock.Anything).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid email",
			req: domain.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *repository.MockUserRepository, tokenRepo *repository.MockTokenRepository) {
				userRepo.On("GetByEmail", context.Background(), "nonexistent@example.com").Return(nil, repository.ErrNotFound)
			},
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			req: domain.LoginRequest{
				Email:    userEmail,
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *repository.MockUserRepository, tokenRepo *repository.MockTokenRepository) {
				userRepo.On("GetByEmail", context.Background(), userEmail).Return(testUser, nil)
			},
			expectedErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(repository.MockUserRepository)
			tokenRepo := new(repository.MockTokenRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, tokenRepo)
			}

			svc := &authService{
				userRepo:  userRepo,
				tokenRepo: tokenRepo,
				txManager: nil,
				logger:    logger,
				cfg: AuthConfig{
					AccessSecret:    "test-secret",
					AccessTokenTTL:  time.Hour,
					RefreshTokenTTL: time.Hour * 24 * 7,
				},
			}

			result, err := svc.Login(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
			}

			userRepo.AssertExpectations(t)
			tokenRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()

	tests := []struct {
		name        string
		rawToken    string
		setupMocks  func(*repository.MockTokenRepository)
		expectedErr error
	}{
		{
			name:     "success",
			rawToken: "valid-refresh-token",
			setupMocks: func(tokenRepo *repository.MockTokenRepository) {
				tokenRepo.On("Revoke", context.Background(), mock.Anything).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:     "invalid token - not found",
			rawToken: "invalid-refresh-token",
			setupMocks: func(tokenRepo *repository.MockTokenRepository) {
				tokenRepo.On("Revoke", context.Background(), mock.Anything).Return(repository.ErrNotFound)
			},
			expectedErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := new(repository.MockTokenRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(tokenRepo)
			}

			svc := &authService{
				userRepo:  nil,
				tokenRepo: tokenRepo,
				txManager: nil,
				logger:    logger,
				cfg:       AuthConfig{},
			}

			err := svc.Logout(context.Background(), tt.rawToken)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			tokenRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()

	userID := ksuid.New().String()
	now := time.Now()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
		TenantID: ksuid.New().String(),
		Role:     domain.RoleOwner,
	}
	testToken := &domain.RefreshToken{
		ID:        ksuid.New().String(),
		UserID:    userID,
		TokenHash: "hashed-token",
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}
	expiredToken := &domain.RefreshToken{
		ID:        ksuid.New().String(),
		UserID:    userID,
		TokenHash: "hashed-expired-token",
		ExpiresAt: now.Add(-time.Hour),
		CreatedAt: now.Add(-time.Hour * 2),
	}
	revokedToken := &domain.RefreshToken{
		ID:        ksuid.New().String(),
		UserID:    userID,
		TokenHash: "hashed-revoked-token",
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
		RevokedAt: &now,
	}

	tests := []struct {
		name        string
		rawToken    string
		setupMocks  func(*repository.MockTokenRepository, *repository.MockUserRepository, *txmanager.MockTxManager)
		expectedErr error
	}{
		{
			name:     "success",
			rawToken: "valid-refresh-token",
			setupMocks: func(tokenRepo *repository.MockTokenRepository, userRepo *repository.MockUserRepository, txManager *txmanager.MockTxManager) {
				tokenRepo.On("GetByHash", context.Background(), mock.Anything).Return(testToken, nil)
				userRepo.On("GetByID", context.Background(), userID).Return(testUser, nil)
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(context.Background())
				})
				tokenRepo.On("Revoke", context.Background(), mock.Anything).Return(nil)
				tokenRepo.On("Store", context.Background(), mock.Anything).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:     "invalid token - not found",
			rawToken: "invalid-refresh-token",
			setupMocks: func(tokenRepo *repository.MockTokenRepository, userRepo *repository.MockUserRepository, txManager *txmanager.MockTxManager) {
				tokenRepo.On("GetByHash", context.Background(), mock.Anything).Return(nil, repository.ErrNotFound)
			},
			expectedErr: ErrInvalidToken,
		},
		{
			name:     "expired token",
			rawToken: "expired-refresh-token",
			setupMocks: func(tokenRepo *repository.MockTokenRepository, userRepo *repository.MockUserRepository, txManager *txmanager.MockTxManager) {
				tokenRepo.On("GetByHash", context.Background(), mock.Anything).Return(expiredToken, nil)
			},
			expectedErr: ErrInvalidToken,
		},
		{
			name:     "revoked token - should revoke all",
			rawToken: "revoked-refresh-token",
			setupMocks: func(tokenRepo *repository.MockTokenRepository, userRepo *repository.MockUserRepository, txManager *txmanager.MockTxManager) {
				tokenRepo.On("GetByHash", context.Background(), mock.Anything).Return(revokedToken, nil)
				tokenRepo.On("RevokeAllForUser", context.Background(), userID).Return(nil)
			},
			expectedErr: ErrInvalidToken,
		},
		{
			name:     "user not found",
			rawToken: "valid-refresh-token",
			setupMocks: func(tokenRepo *repository.MockTokenRepository, userRepo *repository.MockUserRepository, txManager *txmanager.MockTxManager) {
				tokenRepo.On("GetByHash", context.Background(), mock.Anything).Return(testToken, nil)
				userRepo.On("GetByID", context.Background(), userID).Return(nil, repository.ErrNotFound)
			},
			expectedErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := new(repository.MockTokenRepository)
			userRepo := new(repository.MockUserRepository)
			txManager := new(txmanager.MockTxManager)

			if tt.setupMocks != nil {
				tt.setupMocks(tokenRepo, userRepo, txManager)
			}

			svc := &authService{
				userRepo:  userRepo,
				tokenRepo: tokenRepo,
				txManager: txManager,
				logger:    logger,
				cfg: AuthConfig{
					AccessSecret:    "test-secret",
					AccessTokenTTL:  time.Hour,
					RefreshTokenTTL: time.Hour * 24 * 7,
				},
			}

			result, err := svc.Refresh(context.Background(), tt.rawToken)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
			}

			tokenRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}
