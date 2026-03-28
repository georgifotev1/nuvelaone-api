package repository

import (
	"context"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvitationRepository interface {
	GetByID(ctx context.Context, id string) (*domain.UserInvitation, error)
	GetByToken(ctx context.Context, token string) (*domain.UserInvitation, error)
	GetByEmailAndTenant(ctx context.Context, email, tenantID string) (*domain.UserInvitation, error)
	ListByTenant(ctx context.Context, tenantID string) ([]domain.UserInvitation, error)
	Create(ctx context.Context, invitation *domain.UserInvitation) error
	Update(ctx context.Context, invitation *domain.UserInvitation) error
	Delete(ctx context.Context, id string) error
}

type invitationRepository struct {
	pool *pgxpool.Pool
}

func NewInvitationRepository(pool *pgxpool.Pool) InvitationRepository {
	return &invitationRepository{
		pool: pool,
	}
}

func (r *invitationRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *invitationRepository) GetByID(ctx context.Context, id string) (*domain.UserInvitation, error) {
	query := `
		SELECT id, email, name, phone, token, role, invited_by, expires_at, tenant_id, accepted, created_at
		FROM user_invitations WHERE id = $1`

	var inv domain.UserInvitation
	row := r.dbFromContext(ctx).QueryRow(ctx, query, id)
	err := row.Scan(
		&inv.ID,
		&inv.Email,
		&inv.Name,
		&inv.Phone,
		&inv.Token,
		&inv.Role,
		&inv.InvitedBy,
		&inv.ExpiresAt,
		&inv.TenantID,
		&inv.Accepted,
		&inv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitationRepository.GetByID: %w", MapError(err))
	}

	return &inv, nil
}

func (r *invitationRepository) GetByToken(ctx context.Context, token string) (*domain.UserInvitation, error) {
	query := `
		SELECT id, email, name, phone, token, role, invited_by, expires_at, tenant_id, accepted, created_at
		FROM user_invitations WHERE token = $1`

	var inv domain.UserInvitation
	row := r.dbFromContext(ctx).QueryRow(ctx, query, token)
	err := row.Scan(
		&inv.ID,
		&inv.Email,
		&inv.Name,
		&inv.Phone,
		&inv.Token,
		&inv.Role,
		&inv.InvitedBy,
		&inv.ExpiresAt,
		&inv.TenantID,
		&inv.Accepted,
		&inv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitationRepository.GetByToken: %w", MapError(err))
	}

	return &inv, nil
}

func (r *invitationRepository) GetByEmailAndTenant(ctx context.Context, email, tenantID string) (*domain.UserInvitation, error) {
	query := `
		SELECT id, email, name, phone, token, role, invited_by, expires_at, tenant_id, accepted, created_at
		FROM user_invitations WHERE email = $1 AND tenant_id = $2 AND accepted = FALSE`

	var inv domain.UserInvitation
	row := r.dbFromContext(ctx).QueryRow(ctx, query, email, tenantID)
	err := row.Scan(
		&inv.ID,
		&inv.Email,
		&inv.Name,
		&inv.Phone,
		&inv.Token,
		&inv.Role,
		&inv.InvitedBy,
		&inv.ExpiresAt,
		&inv.TenantID,
		&inv.Accepted,
		&inv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitationRepository.GetByEmailAndTenant: %w", MapError(err))
	}

	return &inv, nil
}

func (r *invitationRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.UserInvitation, error) {
	query := `
		SELECT id, email, name, phone, token, role, invited_by, expires_at, tenant_id, accepted, created_at
		FROM user_invitations WHERE tenant_id = $1 AND accepted = FALSE ORDER BY created_at DESC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("invitationRepository.ListByTenant: %w", MapError(err))
	}
	defer rows.Close()

	invitations := make([]domain.UserInvitation, 0)
	for rows.Next() {
		var inv domain.UserInvitation
		if err := rows.Scan(
			&inv.ID,
			&inv.Email,
			&inv.Name,
			&inv.Phone,
			&inv.Token,
			&inv.Role,
			&inv.InvitedBy,
			&inv.ExpiresAt,
			&inv.TenantID,
			&inv.Accepted,
			&inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("invitationRepository.ListByTenant scan: %w", err)
		}
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

func (r *invitationRepository) Create(ctx context.Context, invitation *domain.UserInvitation) error {
	query := `
		INSERT INTO user_invitations (id, email, name, phone, token, role, invited_by, expires_at, tenant_id, accepted, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`

	err := r.dbFromContext(ctx).QueryRow(ctx, query,
		invitation.ID,
		invitation.Email,
		invitation.Name,
		invitation.Phone,
		invitation.Token,
		invitation.Role,
		invitation.InvitedBy,
		invitation.ExpiresAt,
		invitation.TenantID,
		invitation.Accepted,
		invitation.CreatedAt,
	).Scan(&invitation.CreatedAt)
	if err != nil {
		return fmt.Errorf("invitationRepository.Create: %w", MapError(err))
	}

	return nil
}

func (r *invitationRepository) Update(ctx context.Context, invitation *domain.UserInvitation) error {
	query := `
		UPDATE user_invitations 
		SET token = $2, expires_at = $3, accepted = $4
		WHERE id = $1`

	result, err := r.dbFromContext(ctx).Exec(ctx, query,
		invitation.ID,
		invitation.Token,
		invitation.ExpiresAt,
		invitation.Accepted,
	)
	if err != nil {
		return fmt.Errorf("invitationRepository.Update: %w", MapError(err))
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *invitationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM user_invitations WHERE id = $1`
	result, err := r.dbFromContext(ctx).Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("invitationRepository.Delete: %w", MapError(err))
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
