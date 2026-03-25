package repository

import (
	"context"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Store(ctx context.Context, token *domain.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockTokenRepository) Revoke(ctx context.Context, hash string) error {
	args := m.Called(ctx, hash)
	return args.Error(0)
}

func (m *MockTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
