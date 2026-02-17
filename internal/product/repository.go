package product

import (
	"context"
	"errors"

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
	_, err := r.dbPool.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, product *Product) error {
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

	cmd, err := r.dbPool.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("failed save version conflict")
	}
	return nil
}

func (r *Repository) Load() error {
	return nil
}
