package user

import (
	"context"
	"shop-api/internal/database"
	"shop-api/internal/errs"

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

func (r *Repository) Create(ctx context.Context, user *User) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	INSERT INTO users(id, email, password_hash, role)
	VALUES(@id, @email, @password_hash, @role)
	`
	args := pgx.NamedArgs{
		"id":            user.id,
		"email":         user.email.value,
		"password_hash": user.passwordHash.value,
		"role":          user.role,
	}

	_, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) Save(ctx context.Context, user *User) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	UPDATE users
	SET
		email = @email,
		password_hash = @password_hash,
		role = @role,
		email_verified = @email_verified,
		last_login_at = @last_login_at,
		updated_at = @updated_at,
		deleted_at = @deleted_at,
		version = version + 1
	WHERE id = @id AND version = @version AND deleted_at IS NULL
	`
	args := pgx.NamedArgs{
		"id":             user.id,
		"email":          user.email.value,
		"password_hash":  user.passwordHash.value,
		"role":           user.role,
		"email_verified": user.emailVerified,
		"last_login_at":  user.lastLoginAt,
		"updated_at":     user.updatedAt,
		"deleted_at":     user.deletedAt,
		"version":        user.version,
	}

	row, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errs.ErrVersionConflict
	}
	user.version++
	return nil
}

func (r *Repository) GetByEmail(ctx context.Context, email Email) (*User, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		SELECT id, email, password_hash, role, is_active, email_verified, last_login_at, created_at, updated_at, deleted_at, version
		FROM users
		WHERE email = @email AND deleted_at IS NULL
	`

	args := pgx.NamedArgs{
		"email": email.value,
	}

	row := exec.QueryRow(ctx, query, args)

	u := &User{
		passwordHash: PasswordHash{},
		email:        Email{},
	}
	if err := row.Scan(
		&u.id,
		&u.email.value,
		&u.passwordHash.value,
		&u.role,
		&u.isActive,
		&u.emailVerified,
		&u.lastLoginAt,
		&u.createdAt,
		&u.updatedAt,
		&u.deletedAt,
		&u.version,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}
