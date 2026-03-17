package repository

import (
	"context"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("userRepository.GetByID: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("userRepository.GetByEmail: %w", err)
	}
	return user, nil
}

func (r *userRepository) List(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM users ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("userRepository.List: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("userRepository.List scan: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, user.Email, user.Password)
	return row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET email = $1, updated_at = NOW()
		WHERE id = $2`
	_, err := r.db.Exec(ctx, query, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("userRepository.Delete: %w", err)
	}
	return nil
}
