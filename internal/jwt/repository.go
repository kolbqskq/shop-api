package jwt

import (
	"context"
	"shop-api/internal/database"
	"shop-api/internal/errs"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	dbPool *pgxpool.Pool
}

type RepositoryDeps struct {
	DbPool *pgxpool.Pool
}

func NewRepository(deps RepositoryDeps) *Repository {
	return &Repository{
		dbPool: deps.DbPool,
	}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	INSERT INTO refresh_tokens (user_id, token, expires_at)
	VALUES (@user_id, @token, @expires_at)
	`

	args := pgx.NamedArgs{
		"user_id":    userID,
		"token":      token,
		"expires_at": exp,
	}

	if _, err := exec.Exec(ctx, query, args); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Validate(ctx context.Context, token string) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	SELECT EXISTS (
		SELECT 1
		FROM refresh_tokens
		WHERE token = @token AND expires_at > NOW()
	)
	`
	args := pgx.NamedArgs{
		"token": token,
	}
	var exists bool
	if err := exec.QueryRow(ctx, query, args).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		return errs.ErrTokenNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, token string) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		DELETE FROM refresh_tokens WHERE token = @token
	`
	args := pgx.NamedArgs{
		"token": token,
	}

	row, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	if row.RowsAffected() == 0 {
		return errs.ErrTokenNotFound
	}
	return nil
}

func (r *Repository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		DELETE FROM refresh_tokens WHERE user_id = @user_id
	`
	args := pgx.NamedArgs{
		"user_id": userID,
	}

	_, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	return nil
}
