package repository

import (
	"context"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepository interface {
	Store(ctx context.Context, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, hash string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type tokenRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenRepository{pool: pool}
}

func (r *tokenRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *tokenRepository) Store(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.dbFromContext(ctx).Exec(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	return MapError(err)
}

func (r *tokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens WHERE token_hash = $1`
	row := r.dbFromContext(ctx).QueryRow(ctx, query, hash)
	t := &domain.RefreshToken{}
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt)
	if err != nil {
		return nil, MapError(err)
	}
	return t, nil
}

func (r *tokenRepository) Revoke(ctx context.Context, hash string) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL`
	result, err := r.dbFromContext(ctx).Exec(ctx, query, hash)
	if err != nil {
		return MapError(err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *tokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := r.dbFromContext(ctx).Exec(ctx, query, userID)
	return MapError(err)
}

func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.pool.Exec(ctx, query)
	return MapError(err)
}
