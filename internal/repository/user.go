package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/georgifotev1/nuvelaone-api/internal/cache"
	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db    *pgxpool.Pool
	cache *cache.UserStore
}

func NewUserRepository(db *pgxpool.Pool, cache *cache.UserStore) UserRepository {
	return &userRepository{
		db:    db,
		cache: cache,
	}
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := r.cache.Get(ctx, id)
	if err != nil {
		log.Printf("cache get failed for user:%s: %v", id, err)
	}
	if user != nil {
		return user, nil
	}

	user = &domain.User{}
	query := `
		SELECT id, email, password, name, phone, tenant_id, avatar, role, verified, created_at, updated_at
		FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	err = row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
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
		return nil, fmt.Errorf("userRepository.GetByID: %w", err)
	}

	if err := r.cache.Set(ctx, id, user); err != nil {
		log.Printf("cache set failed for user:%s: %v", id, err)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT id, email, password, name, phone, tenant_id, avatar, role, verified, created_at, updated_at
		FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Phone, &user.TenantID, &user.Avatar, &user.Role, &user.Verified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("userRepository.GetByEmail: %w", err)
	}
	return user, nil
}

func (r *userRepository) List(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, email, password, name, phone, tenant_id, avatar, role, verified, created_at, updated_at
		FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("userRepository.List: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.Phone, &u.TenantID, &u.Avatar, &u.Role, &u.Verified, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("userRepository.List scan: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password, name, phone, tenant_id, avatar, role, verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`
	row := r.db.QueryRow(ctx, query, user.ID, user.Email, user.Password, user.Name, user.Phone, user.TenantID, user.Avatar, user.Role, user.Verified)
	return row.Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET email = $1, name = $2, phone = $3, avatar = $4, role = $5, verified = $6, updated_at = NOW()
		WHERE id = $7`
	_, err := r.db.Exec(ctx, query, user.Email, user.Name, user.Phone, user.Avatar, user.Role, user.Verified, user.ID)
	if err != nil {
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	if err := r.cache.Delete(ctx, user.ID); err != nil {
		log.Printf("cache delete failed for user:%s: %v", user.ID, err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("userRepository.Delete: %w", err)
	}
	if err := r.cache.Delete(ctx, id); err != nil {
		log.Printf("cache delete failed for user:%s: %v", id, err)
	}
	return nil
}
