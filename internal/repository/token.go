package repository

import (
	"context"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepository interface {
	Store(ctx context.Context, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, hash string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	RevokeAllForCustomer(ctx context.Context, customerID string) error
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
		INSERT INTO refresh_tokens (id, entity_id, entity_type, tenant_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.dbFromContext(ctx).Exec(ctx, query,
		token.ID, token.EntityID, token.EntityType, token.TenantID,
		token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *tokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, entity_id, entity_type, tenant_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens WHERE token_hash = $1`
	row := r.dbFromContext(ctx).QueryRow(ctx, query, hash)
	t := &domain.RefreshToken{}
	err := row.Scan(&t.ID, &t.EntityID, &t.EntityType, &t.TenantID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("tokenRepository.GetByHash: %w", apperr.NotFound("token not found", err))
		}
		return nil, fmt.Errorf("tokenRepository.GetByHash: %w", apperr.Internal(err))
	}
	return t, nil
}

func (r *tokenRepository) Revoke(ctx context.Context, hash string) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL`
	result, err := r.dbFromContext(ctx).Exec(ctx, query, hash)
	if err != nil {
		return apperr.Internal(err)
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("token not found", nil)
	}
	return nil
}

func (r *tokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE entity_id = $1 AND entity_type = 'user' AND revoked_at IS NULL`
	_, err := r.dbFromContext(ctx).Exec(ctx, query, userID)
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *tokenRepository) RevokeAllForCustomer(ctx context.Context, customerID string) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE entity_id = $1 AND entity_type = 'customer' AND revoked_at IS NULL`
	_, err := r.dbFromContext(ctx).Exec(ctx, query, customerID)
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}
