package repository

import (
	"context"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	Create(ctx context.Context, tenant *domain.Tenant) error
	Update(ctx context.Context, tenant *domain.Tenant) error
}

type tenantRepository struct {
	pool *pgxpool.Pool
}

func NewTenantRepository(pool *pgxpool.Pool) TenantRepository {
	return &tenantRepository{
		pool: pool,
	}
}

func (r *tenantRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *tenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, phone, email, address_id, created_at, updated_at
		FROM tenants WHERE id = $1`

	var tenant domain.Tenant
	row := r.dbFromContext(ctx).QueryRow(ctx, query, id)
	err := row.Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Phone,
		&tenant.Email,
		&tenant.AddressID,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("tenantRepository.GetByID: %w", MapError(err))
	}

	return &tenant, nil
}

func (r *tenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, phone, email, address_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.dbFromContext(ctx).Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Phone,
		tenant.Email,
		tenant.AddressID,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("tenantRepository.Create: %w", MapError(err))
	}

	return nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		UPDATE tenants 
		SET name = $2, slug = $3, phone = $4, email = $5, address_id = $6, updated_at = $7
		WHERE id = $1`

	result, err := r.dbFromContext(ctx).Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Phone,
		tenant.Email,
		tenant.AddressID,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("tenantRepository.Update: %w", MapError(err))
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
