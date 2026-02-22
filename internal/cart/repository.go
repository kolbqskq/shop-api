package cart

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

func (r *Repository) Create(ctx context.Context, cart *Cart) error {
	exec := database.Executor(ctx, r.dbPool)

	if len(cart.Items) == 0 {
		return errs.ErrEmptyCart
	}

	query :=
		`
	INSERT INTO carts (id, user_id)
	VALUES (@id, @user_id);
	`
	args := pgx.NamedArgs{
		"id":      cart.ID,
		"user_id": cart.UserID,
	}

	_, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddItems(ctx context.Context, cart *Cart) error {
	exec := database.Executor(ctx, r.dbPool)

	if len(cart.Items) == 0 {
		return errs.ErrEmptyCart
	}

	query :=
		`
	INSERT INTO cart_items (cart_id, product_id, quantity)
	VALUES (@cart_id, @product_id, @quantity),
	ON CONFLICT (cart_id, product_id)
	DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity;
	`

	for _, v := range cart.Items {
		args := pgx.NamedArgs{
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
