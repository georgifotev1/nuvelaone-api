package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/pkg/null"
	"github.com/segmentio/ksuid"
)

type CustomerService interface {
	ListByTenant(ctx context.Context, tenantID string) ([]domain.Customer, error)
	GetByID(ctx context.Context, tenantID, id string) (*domain.Customer, error)
	Create(ctx context.Context, tenantID string, req domain.CustomerRequest) (*domain.Customer, error)
	Update(ctx context.Context, tenantID, customerID string, req domain.CustomerRequest) (*domain.Customer, error)
	Delete(ctx context.Context, tenantID, customerID string) error
}

type customerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Customer, error) {
	customers, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("customerService.ListByTenant: %w", err)
	}
	return customers, nil
}

func (s *customerService) GetByID(ctx context.Context, tenantID, id string) (*domain.Customer, error) {
	customer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("customerService.GetByID: %w", err)
	}
	return customer, nil
}

func (s *customerService) Create(ctx context.Context, tenantID string, req domain.CustomerRequest) (*domain.Customer, error) {
	now := time.Now()
	customer := &domain.Customer{
		ID:        ksuid.New().String(),
		TenantID:  tenantID,
		Name:      req.Name,
		Email:     null.FromString(req.Email),
		Phone:     req.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("customerService.Create: %w", err)
	}

	return customer, nil
}

func (s *customerService) Update(ctx context.Context, tenantID, customerID string, req domain.CustomerRequest) (*domain.Customer, error) {
	customer, err := s.repo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("customerService.Update get: %w", err)
	}

	customer.Name = req.Name
	customer.Email = null.FromString(req.Email)
	customer.Phone = req.Phone
	customer.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, customer); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrConflict
		}

		return nil, fmt.Errorf("customerService.Update: %w", err)
	}

	return customer, nil
}

func (s *customerService) Delete(ctx context.Context, tenantID, customerID string) error {
	if err := s.repo.Delete(ctx, tenantID, customerID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("customerService.Delete: %w", err)
	}
	return nil
}
