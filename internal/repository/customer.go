package repository

import (
	"context"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Customer, error)
	GetByEmail(ctx context.Context, tenantID, email string) (*domain.Customer, error)
	ListByTenant(ctx context.Context, tenantID string) ([]domain.Customer, error)
	Create(ctx context.Context, customer *domain.Customer) error
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, tenantID, id string) error
}

type customerRepository struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) CustomerRepository {
	return &customerRepository{pool: pool}
}

func (r *customerRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *customerRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Customer, error) {
	query := `
		SELECT id, tenant_id, name, email, phone, created_at, updated_at
		FROM customers WHERE id = $1 AND tenant_id = $2`

	var c domain.Customer
	err := r.dbFromContext(ctx).QueryRow(ctx, query, id, tenantID).Scan(
		&c.ID,
		&c.TenantID,
		&c.Name,
		&c.Email,
		&c.Phone,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("customerRepository.GetByID: %w", apperr.NotFound("customer not found", err))
		}
		return nil, fmt.Errorf("customerRepository.GetByID: %w", apperr.Internal(err))
	}
	return &c, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, tenantID, email string) (*domain.Customer, error) {
	query := `
		SELECT id, tenant_id, name, email, password, phone, created_at, updated_at
		FROM customers WHERE LOWER(email) = LOWER($1) AND tenant_id = $2`

	var c domain.Customer
	err := r.dbFromContext(ctx).QueryRow(ctx, query, email, tenantID).Scan(
		&c.ID,
		&c.TenantID,
		&c.Name,
		&c.Email,
		&c.Password,
		&c.Phone,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("customerRepository.GetByEmail: %w", apperr.NotFound("customer not found", err))
		}
		return nil, fmt.Errorf("customerRepository.GetByEmail: %w", apperr.Internal(err))
	}
	return &c, nil
}

func (r *customerRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Customer, error) {
	query := `
		SELECT id, tenant_id, name, email, phone, created_at, updated_at
		FROM customers WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.ListByTenant: %w", apperr.Internal(err))
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0)
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(
			&c.ID,
			&c.TenantID,
			&c.Name,
			&c.Email,
			&c.Phone,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("customerRepository.ListByTenant scan: %w", apperr.Internal(err))
		}
		customers = append(customers, c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("customerRepository.ListByTenant: %w", apperr.Internal(rows.Err()))
	}
	return customers, nil
}

func (r *customerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	query := `
		INSERT INTO customers (id, tenant_id, name, email, password, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.dbFromContext(ctx).Exec(ctx, query,
		customer.ID,
		customer.TenantID,
		customer.Name,
		customer.Email,
		customer.Password,
		customer.Phone,
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("customerRepository.Create: %w", apperr.Conflict("customer with that email already exists"))
		}
		return fmt.Errorf("customerRepository.Create: %w", apperr.Internal(err))
	}
	return nil
}

func (r *customerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	query := `
		UPDATE customers
		SET name = $2, email = $3, phone = $4, password = COALESCE($5, password), updated_at = $6
		WHERE id = $1`

	result, err := r.dbFromContext(ctx).Exec(ctx, query,
		customer.ID,
		customer.Name,
		customer.Email,
		customer.Phone,
		customer.Password,
		customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("customerRepository.Update: %w", apperr.Internal(err))
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("customer not found", nil)
	}
	return nil
}

func (r *customerRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM customers WHERE id = $1 AND tenant_id = $2`
	result, err := r.dbFromContext(ctx).Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("customerRepository.Delete: %w", apperr.Internal(err))
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("customer not found", nil)
	}
	return nil
}
