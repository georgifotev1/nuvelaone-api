package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/cache"
	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/ksuid"
)

type ServiceRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Service, error)
	ListByTenant(ctx context.Context, tenantID string) ([]domain.Service, error)
	ListVisible(ctx context.Context, tenantID string) ([]domain.Service, error)
	Create(ctx context.Context, service *domain.Service) error
	Update(ctx context.Context, service *domain.Service) error
	Delete(ctx context.Context, tenantID, id string) error
	AssignUsers(ctx context.Context, serviceID string, userIDs []string, tenantID string) error
	GetUserServices(ctx context.Context, serviceID string) ([]domain.UserService, error)
	GetProvidersByService(ctx context.Context, tenantID, serviceID string) ([]domain.User, error)
	GetProviderIDsByService(ctx context.Context, tenantID, serviceID string) ([]string, error)
}

type serviceRepository struct {
	pool  *pgxpool.Pool
	cache *cache.ServiceStore
}

func NewServiceRepository(pool *pgxpool.Pool, cache *cache.ServiceStore) ServiceRepository {
	return &serviceRepository{pool: pool, cache: cache}
}

func (r *serviceRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *serviceRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Service, error) {
	service, err := r.cache.Get(ctx, id)
	if err != nil {
		fmt.Printf("cache get failed for service:%s: %v\n", id, err)
	}
	if service != nil && service.TenantID == tenantID {
		return service, nil
	}

	query := `
		SELECT id, title, description, duration, buffer, cost, visible, tenant_id, created_at, updated_at
		FROM services WHERE id = $1 AND tenant_id = $2`

	var s domain.Service
	err = r.dbFromContext(ctx).QueryRow(ctx, query, id, tenantID).Scan(
		&s.ID,
		&s.Title,
		&s.Description,
		&s.Duration,
		&s.Buffer,
		&s.Cost,
		&s.Visible,
		&s.TenantID,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("serviceRepository.GetByID: %w", apperr.NotFound("service not found", err))
		}
		return nil, fmt.Errorf("serviceRepository.GetByID: %w", apperr.Internal(err))
	}

	if err := r.cache.Set(ctx, id, &s); err != nil {
		fmt.Printf("cache set failed for service:%s: %v\n", id, err)
	}

	return &s, nil
}

func (r *serviceRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Service, error) {
	query := `
		SELECT id, title, description, duration, buffer, cost, visible, tenant_id, created_at, updated_at
		FROM services WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("serviceRepository.ListByTenant: %w", apperr.Internal(err))
	}
	defer rows.Close()

	services := make([]domain.Service, 0)
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Description,
			&s.Duration,
			&s.Buffer,
			&s.Cost,
			&s.Visible,
			&s.TenantID,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("serviceRepository.ListByTenant scan: %w", apperr.Internal(err))
		}
		services = append(services, s)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("serviceRepository.ListByTenant: %w", apperr.Internal(rows.Err()))
	}
	return services, nil
}

func (r *serviceRepository) ListVisible(ctx context.Context, tenantID string) ([]domain.Service, error) {
	query := `
		SELECT s.id, s.title, s.description, s.duration, s.buffer, s.cost, s.visible, s.tenant_id, s.created_at, s.updated_at
		FROM services s
		INNER JOIN user_services us ON s.id = us.service_id
		WHERE s.tenant_id = $1 AND s.visible = true
		GROUP BY s.id
		ORDER BY s.created_at DESC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("serviceRepository.ListVisible: %w", apperr.Internal(err))
	}
	defer rows.Close()

	services := make([]domain.Service, 0)
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Description,
			&s.Duration,
			&s.Buffer,
			&s.Cost,
			&s.Visible,
			&s.TenantID,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("serviceRepository.ListVisible scan: %w", apperr.Internal(err))
		}
		services = append(services, s)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("serviceRepository.ListVisible: %w", apperr.Internal(rows.Err()))
	}
	return services, nil
}

func (r *serviceRepository) Create(ctx context.Context, service *domain.Service) error {
	query := `
		INSERT INTO services (id, title, description, duration, buffer, cost, visible, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.dbFromContext(ctx).Exec(ctx, query,
		service.ID,
		service.Title,
		service.Description,
		service.Duration,
		service.Buffer,
		service.Cost,
		service.Visible,
		service.TenantID,
		service.CreatedAt,
		service.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("serviceRepository.Create: %w", apperr.Conflict("service already exists"))
		}
		return fmt.Errorf("serviceRepository.Create: %w", apperr.Internal(err))
	}
	return nil
}

func (r *serviceRepository) Update(ctx context.Context, service *domain.Service) error {
	query := `
		UPDATE services
		SET title = $2, description = $3, duration = $4, buffer = $5, cost = $6, visible = $7, updated_at = $8
		WHERE id = $1`

	result, err := r.dbFromContext(ctx).Exec(ctx, query,
		service.ID,
		service.Title,
		service.Description,
		service.Duration,
		service.Buffer,
		service.Cost,
		service.Visible,
		service.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("serviceRepository.Update: %w", apperr.Internal(err))
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("service not found", nil)
	}

	if err := r.cache.Delete(ctx, service.ID); err != nil {
		fmt.Printf("cache delete failed for service:%s: %v\n", service.ID, err)
	}

	return nil
}

func (r *serviceRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM services WHERE id = $1 AND tenant_id = $2`
	result, err := r.dbFromContext(ctx).Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("serviceRepository.Delete: %w", apperr.Internal(err))
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("service not found", nil)
	}

	if err := r.cache.Delete(ctx, id); err != nil {
		fmt.Printf("cache delete failed for service:%s: %v\n", id, err)
	}

	return nil
}

func (r *serviceRepository) AssignUsers(ctx context.Context, serviceID string, userIDs []string, tenantID string) error {
	if len(userIDs) == 0 {
		return nil
	}

	db := r.dbFromContext(ctx)

	deleteQuery := `DELETE FROM user_services WHERE service_id = $1 AND tenant_id = $2`
	if _, err := db.Exec(ctx, deleteQuery, serviceID, tenantID); err != nil {
		return fmt.Errorf("serviceRepository.AssignUsers delete: %w", apperr.Internal(err))
	}

	insertQuery := `
		INSERT INTO user_services (id, user_id, service_id, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for _, userID := range userIDs {
		now := time.Now()
		if _, err := db.Exec(ctx, insertQuery,
			ksuid.New().String(),
			userID,
			serviceID,
			tenantID,
			now,
			now,
		); err != nil {
			return fmt.Errorf("serviceRepository.AssignUsers insert: %w", apperr.Internal(err))
		}
	}

	return nil
}

func (r *serviceRepository) GetUserServices(ctx context.Context, serviceID string) ([]domain.UserService, error) {
	query := `
		SELECT id, user_id, service_id, tenant_id, created_at, updated_at
		FROM user_services WHERE service_id = $1`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("serviceRepository.GetUserServices: %w", apperr.Internal(err))
	}
	defer rows.Close()

	var userServices []domain.UserService
	for rows.Next() {
		var us domain.UserService
		if err := rows.Scan(&us.ID, &us.UserID, &us.ServiceID, &us.TenantID, &us.CreatedAt, &us.UpdatedAt); err != nil {
			return nil, fmt.Errorf("serviceRepository.GetUserServices scan: %w", apperr.Internal(err))
		}
		userServices = append(userServices, us)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("serviceRepository.GetUserServices: %w", apperr.Internal(rows.Err()))
	}
	return userServices, nil
}

func (r *serviceRepository) GetProvidersByService(ctx context.Context, tenantID, serviceID string) ([]domain.User, error) {
	query := `
		SELECT u.id, u.email, u.password, u.name, u.phone, u.tenant_id, u.avatar, u.role, u.verified, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_services us ON u.id = us.user_id
		WHERE us.service_id = $1 AND u.tenant_id = $2
		ORDER BY u.name ASC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, serviceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("serviceRepository.GetProvidersByService: %w", apperr.Internal(err))
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Password,
			&u.Name,
			&u.Phone,
			&u.TenantID,
			&u.Avatar,
			&u.Role,
			&u.Verified,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("serviceRepository.GetProvidersByService scan: %w", apperr.Internal(err))
		}
		users = append(users, u)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("serviceRepository.GetProvidersByService: %w", apperr.Internal(rows.Err()))
	}
	return users, nil
}

func (r *serviceRepository) GetProviderIDsByService(ctx context.Context, tenantID, serviceID string) ([]string, error) {
	query := `
		SELECT us.user_id FROM user_services us
		INNER JOIN users u ON u.id = us.user_id
		WHERE us.service_id = $1 AND us.tenant_id = $2
		ORDER BY u.name ASC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, serviceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("serviceRepository.GetProviderIDsByService: %w", apperr.Internal(err))
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("serviceRepository.GetProviderIDsByService scan: %w", apperr.Internal(err))
		}
		ids = append(ids, id)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("serviceRepository.GetProviderIDsByService: %w", apperr.Internal(rows.Err()))
	}
	return ids, nil
}
