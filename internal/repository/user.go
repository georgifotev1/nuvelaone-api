package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/georgifotev1/nuvelaone-api/internal/cache"
	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, tenantID string) ([]domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	pool  *pgxpool.Pool
	cache *cache.UserStore
}

func NewUserRepository(pool *pgxpool.Pool, cache *cache.UserStore) UserRepository {
	return &userRepository{
		pool:  pool,
		cache: cache,
	}
}

func (r *userRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := r.cache.Get(ctx, id)
	if err != nil {
		log.Printf("cache get failed for user:%s: %v", id, err)
	}
	if user != nil {
		return user, nil
	}

	query := `
		SELECT id, email, name, phone, tenant_id, avatar, role, verified, created_at, updated_at
		FROM users WHERE id = $1`

	row := r.dbFromContext(ctx).QueryRow(ctx, query, id)
	user = &domain.User{}
	err = row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Phone,
		&user.TenantID,
		&user.Avatar,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("userRepository.GetByID: %w", apperr.NotFound("user not found", err))
		}
		return nil, fmt.Errorf("userRepository.GetByID: %w", apperr.Internal(err))
	}

	if err := r.cache.Set(ctx, id, user); err != nil {
		log.Printf("cache set failed for user:%s: %v", id, err)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password, name, phone, tenant_id, avatar, role, verified, created_at, updated_at
		FROM users WHERE email = $1`
	row := r.dbFromContext(ctx).QueryRow(ctx, query, email)
	user := &domain.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Phone, &user.TenantID, &user.Avatar, &user.Role, &user.Verified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("userRepository.GetByEmail: %w", apperr.NotFound("user not found", err))
		}
		return nil, fmt.Errorf("userRepository.GetByEmail: %w", apperr.Internal(err))
	}
	return user, nil
}

func (r *userRepository) List(ctx context.Context, tenantID string) ([]domain.User, error) {
	query := `
		SELECT id, email, password, name, phone, tenant_id, avatar, role, verified, created_at, updated_at
		FROM users WHERE tenant_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("userRepository.List: %w", apperr.Internal(err))
	}
	defer rows.Close()

	users := make([]domain.User, 0)
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
			return nil, fmt.Errorf("userRepository.List scan: %w", apperr.Internal(err))
		}
		users = append(users, u)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("userRepository.List: %w", apperr.Internal(rows.Err()))
	}
	return users, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password, name, phone, tenant_id, avatar, role, verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`
	row := r.dbFromContext(ctx).QueryRow(ctx, query, user.ID, user.Email, user.Password, user.Name, user.Phone, user.TenantID, user.Avatar, user.Role, user.Verified)
	err := row.Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("userRepository.Create: %w", apperr.Conflict("user already exists"))
		}
		return fmt.Errorf("userRepository.Create: %w", apperr.Internal(err))
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET email = $1, name = $2, phone = $3, avatar = $4, role = $5, verified = $6, updated_at = NOW()
		WHERE id = $7`
	result, err := r.dbFromContext(ctx).Exec(ctx, query, user.Email, user.Name, user.Phone, user.Avatar, user.Role, user.Verified, user.ID)
	if err != nil {
		return fmt.Errorf("userRepository.Update: %w", apperr.Internal(err))
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("user not found", nil)
	}
	if err := r.cache.Delete(ctx, user.ID); err != nil {
		log.Printf("cache delete failed for user:%s: %v", user.ID, err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.dbFromContext(ctx).Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("userRepository.Delete: %w", apperr.Internal(err))
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("user not found", nil)
	}
	if err := r.cache.Delete(ctx, id); err != nil {
		log.Printf("cache delete failed for user:%s: %v", id, err)
	}
	return nil
}
