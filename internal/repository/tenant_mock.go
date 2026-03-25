package repository

import (
	"context"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}
