package product

import (
	"context"
	"errors"
	"shop-api/pkg/transaction"

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

func (r *Repository) Create(ctx context.Context, product *Product) error {
	exec := r.executor(ctx)

	query :=
		`
	INSERT INTO products (id, name, description, category, price, stock, is_active)
	VALUES (@id, @name, @description, @category, @price, @stock, @is_active)
	`
	args := pgx.NamedArgs{
		"id":          product.ID,
		"name":        product.Name,
		"description": product.Description,
		"category":    product.Category,
		"price":       product.Price.Amount,
		"stock":       product.Stock,
		"is_active":   product.IsActive,
	}
	_, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, product *Product) error {
	exec := r.executor(ctx)

	query :=
		`
	UPDATE products
	SET 
		name = @name,
		description = @description,
		category = @category,
		price = @price,
		stock = @stock,
		is_active = @is_active,
		updated_at = NOW(),
		version = version + 1

	WHERE id = @id AND version = @version
	`
	args := pgx.NamedArgs{
		"id":          product.ID,
		"name":        product.Name,
		"description": product.Description,
		"category":    product.Category,
		"price":       product.Price.Amount,
		"stock":       product.Stock,
		"is_active":   product.IsActive,
		"version":     product.Version,
	}

	cmd, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("failed save version conflict")
	}
	return nil
}

func (r *Repository) Reserve(ctx context.Context, products []Reservation) error {
	exec := r.executor(ctx)

	query :=
		`
	UPDATE products
	SET
		reserved = reserved + @quantity,
		version = version + 1
	WHERE id = @id AND is_active = TRUE AND stock - reserved >= @quantity
	`
	for _, v := range products {

		if v.Quantity <= 0 {
			return errors.New("invalid quantity")
		}

		args := pgx.NamedArgs{
			"id":       v.ProductID,
			"quantity": v.Quantity,
		}

		cmd, err := exec.Exec(ctx, query, args)
		if err != nil {
			return err
		}
		if cmd.RowsAffected() == 0 {
			return errors.New("not enough stock")
		}
	}

	return nil

}

func (r *Repository) Commit(ctx context.Context, products []Reservation) error {
	exec := r.executor(ctx)

	query :=
		`
	UPDATE products
	SET
		reserved = reserved - @quantity,
		stock = stock - @quantity,
		version = version +1
	WHERE id = @id AND reserved >= @quantity
	`
	for _, v := range products {

		if v.Quantity <= 0 {
			return errors.New("invalid quantity")
		}

		args := pgx.NamedArgs{
			"id":       v.ProductID,
			"quantity": v.Quantity,
		}
		cmd, err := exec.Exec(ctx, query, args)
		if err != nil {
			return err
		}

		if cmd.RowsAffected() == 0 {
			return errors.New("not enough reserve")
		}
	}
	return nil
}

func (r *Repository) Release(ctx context.Context, products []Reservation) error {
	exec := r.executor(ctx)

	query :=
		`
	UPDATE products
	SET
		reserved = reserved - @quantity,
		version = version +1
	WHERE id = @id AND reserved >= @quantity
	`
	for _, v := range products {

		if v.Quantity <= 0 {
			return errors.New("invalid quantity")
		}

		args := pgx.NamedArgs{
			"id":       v.ProductID,
			"quantity": v.Quantity,
		}
		cmd, err := exec.Exec(ctx, query, args)
		if err != nil {
			return err
		}

		if cmd.RowsAffected() == 0 {
			return errors.New("not enough reserve")
		}
	}
	return nil
}

func (r *Repository) executor(ctx context.Context) transaction.DBTX {
	if tx, ok := transaction.ExtractTx(ctx); ok {
		return tx
	}
	return r.dbPool
}
