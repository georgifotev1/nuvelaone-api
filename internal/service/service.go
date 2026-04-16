package service

import (
	"context"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/segmentio/ksuid"
)

type ServiceService interface {
	ListByTenant(ctx context.Context, tenantID string) ([]domain.Service, error)
	ListVisible(ctx context.Context, tenantID string) ([]domain.Service, error)
	GetByID(ctx context.Context, tenantID, id string) (*domain.Service, error)
	Create(ctx context.Context, tenantID string, req domain.ServiceRequest) (*domain.Service, error)
	Update(ctx context.Context, tenantID, serviceID string, req domain.ServiceRequest) (*domain.Service, error)
	Delete(ctx context.Context, tenantID, serviceID string) error
	GetProviders(ctx context.Context, tenantID, serviceID string) ([]domain.User, error)
}

type serviceService struct {
	repo repository.ServiceRepository
	tx   txmanager.TxManager
}

func NewServiceService(repo repository.ServiceRepository, tx txmanager.TxManager) ServiceService {
	return &serviceService{repo: repo, tx: tx}
}

func normalizeProviderIDs(ids []string) []string {
	if ids == nil {
		return []string{}
	}
	return ids
}

func (s *serviceService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Service, error) {
	services, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("serviceService.ListByTenant: %w", err)
	}

	for i := range services {
		providerIDs, err := s.repo.GetProviderIDsByService(ctx, tenantID, services[i].ID)
		if err != nil {
			return nil, fmt.Errorf("serviceService.ListByTenant providers: %w", err)
		}
		services[i].ProviderIDs = normalizeProviderIDs(providerIDs)
	}

	return services, nil
}

func (s *serviceService) ListVisible(ctx context.Context, tenantID string) ([]domain.Service, error) {
	services, err := s.repo.ListVisible(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("serviceService.ListVisible: %w", err)
	}

	for i := range services {
		providerIDs, err := s.repo.GetProviderIDsByService(ctx, tenantID, services[i].ID)
		if err != nil {
			return nil, fmt.Errorf("serviceService.ListVisible providers: %w", err)
		}
		services[i].ProviderIDs = normalizeProviderIDs(providerIDs)
	}

	return services, nil
}

func (s *serviceService) GetByID(ctx context.Context, tenantID, id string) (*domain.Service, error) {
	service, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("serviceService.GetByID: %w", err)
	}

	providerIDs, err := s.repo.GetProviderIDsByService(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("serviceService.GetByID providers: %w", err)
	}
	service.ProviderIDs = normalizeProviderIDs(providerIDs)

	return service, nil
}

func (s *serviceService) Create(ctx context.Context, tenantID string, req domain.ServiceRequest) (*domain.Service, error) {
	now := time.Now()
	service := &domain.Service{
		ID:          ksuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Duration:    req.Duration,
		Buffer:      req.Buffer,
		Cost:        req.Cost,
		Visible:     req.Visible,
		TenantID:    tenantID,
		CreatedAt:   now,
		UpdatedAt:   now,
		ProviderIDs: normalizeProviderIDs(req.UserIDs),
	}

	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, service); err != nil {
			return fmt.Errorf("serviceService.Create repo: %w", err)
		}
		if len(req.UserIDs) > 0 {
			if err := s.repo.AssignUsers(ctx, service.ID, req.UserIDs, tenantID); err != nil {
				return fmt.Errorf("serviceService.Create assign users: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (s *serviceService) Update(ctx context.Context, tenantID, serviceID string, req domain.ServiceRequest) (*domain.Service, error) {
	var updated *domain.Service

	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		service, err := s.repo.GetByID(ctx, tenantID, serviceID)
		if err != nil {
			return fmt.Errorf("serviceService.Update get: %w", err)
		}

		if req.Title != "" {
			service.Title = req.Title
		}
		if req.Description != "" {
			service.Description = req.Description
		}
		if req.Duration > 0 {
			service.Duration = req.Duration
		}
		if req.Buffer > 0 {
			service.Buffer = req.Buffer
		}
		if req.Cost > 0 {
			service.Cost = req.Cost
		}
		service.Visible = req.Visible
		service.UpdatedAt = time.Now()

		if err := s.repo.Update(ctx, service); err != nil {
			return fmt.Errorf("serviceService.Update repo: %w", err)
		}

		if req.UserIDs != nil {
			if err := s.repo.AssignUsers(ctx, service.ID, req.UserIDs, tenantID); err != nil {
				return fmt.Errorf("serviceService.Update assign users: %w", err)
			}
			service.ProviderIDs = normalizeProviderIDs(req.UserIDs)
		} else {
			providerIDs, err := s.repo.GetProviderIDsByService(ctx, tenantID, serviceID)
			if err != nil {
				return fmt.Errorf("serviceService.Update get providers: %w", err)
			}
			service.ProviderIDs = normalizeProviderIDs(providerIDs)
		}

		updated = service
		return nil
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *serviceService) Delete(ctx context.Context, tenantID, serviceID string) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Delete(ctx, tenantID, serviceID); err != nil {
			return fmt.Errorf("serviceService.Delete: %w", err)
		}
		return nil
	})
}

func (s *serviceService) GetProviders(ctx context.Context, tenantID, serviceID string) ([]domain.User, error) {
	providers, err := s.repo.GetProvidersByService(ctx, tenantID, serviceID)
	if err != nil {
		return nil, fmt.Errorf("serviceService.GetProviders: %w", err)
	}
	return providers, nil
}
