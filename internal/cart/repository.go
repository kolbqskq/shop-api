package cart

import (
	"context"
	"errors"
	"shop-api/internal/database"
	"shop-api/internal/errs"

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

func (r *Repository) Create(ctx context.Context, cart *Cart) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	INSERT INTO carts (id, user_id)
	VALUES (@id, @user_id)
	ON CONFLICT
	DO NOTHING
	RETURNING id;
	`
	args := pgx.NamedArgs{
		"id":      cart.ID,
		"user_id": cart.userID,
	}
	var id uuid.UUID
	err := exec.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrCartAlreadyExist
		}
		return err
	}
	return nil
}

func (r *Repository) Save(ctx context.Context, cart *Cart) error {
	exec := database.Executor(ctx, r.dbPool)
	if _, ok := exec.(pgx.Tx); !ok {
		return errors.New("Save must be called inside transaction")
	}
	query :=
		`
		UPDATE carts
		SET 
			status = @status,
			version = version + 1,
			updated_at = NOW()
		WHERE id = @id AND version = @version;
	`
	args := pgx.NamedArgs{
		"id":      cart.ID,
		"status":  cart.status,
		"version": cart.version,
	}
	row, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errs.ErrVersionConflict
	}
	query =
		`
	DELETE FROM cart_items
	WHERE cart_id = @id;
	`
	_, err = exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if len(cart.items) == 0 {
		return nil
	}
	query =
		`
	INSERT INTO cart_items (cart_id, product_id, quantity)
	SELECT @cart_id, @product_id, @quantity
	`
	for _, v := range cart.items {
		args = pgx.NamedArgs{
			"cart_id":    cart.ID,
			"product_id": v.ProductID,
			"quantity":   v.Quantity,
		}
		_, err := exec.Exec(ctx, query, args)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetActiveCart(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	exec := database.Executor(ctx, r.dbPool)
	query :=
		`
		SELECT id, version
		FROM carts
		WHERE user_id = @user_id AND status = 'active'
	`
	args := pgx.NamedArgs{
		"user_id": userID,
	}
	row := exec.QueryRow(ctx, query, args)
	cart := &Cart{
		userID: userID,
		status: CartStatusActive,
	}
	if err := row.Scan(&cart.id, &cart.version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrCartNotFound
		}
		return nil, err
	}

	query =
		`
	SELECT product_id, quantity
	FROM cart_items
	WHERE cart_id = @cart_id
	ORDER BY added_at;
	`
	args = pgx.NamedArgs{
		"cart_id": cart.ID,
	}
	rows, err := exec.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CartItem
	for rows.Next() {
		var p uuid.UUID
		var q int
		if err := rows.Scan(&p, &q); err != nil {
			return nil, err
		}
		items = append(items, CartItem{
			ProductID: p,
			Quantity:  q,
		})
	}
	cart.items = items
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cart, nil
}
