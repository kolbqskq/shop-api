package product

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

func (r *Repository) Create(ctx context.Context, product *Product) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	INSERT INTO products (id, name, description, category, price, stock, is_active)
	VALUES (@id, @name, @description, @category, @price, @stock, @is_active)
	`
	args := pgx.NamedArgs{
		"id":          product.id,
		"name":        product.name,
		"description": product.description,
		"category":    product.category,
		"price":       product.price.Amount,
		"stock":       product.stock,
		"is_active":   product.isActive,
	}
	_, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) Save(ctx context.Context, product *Product) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	UPDATE products
	SET 
		name = @name,
		description = @description,
		category = @category,
		price = @price,
		stock = @stock,
		reserved = @reserved,
		is_active = @is_active,
		updated_at = NOW(),
		version = version + 1

	WHERE id = @id AND version = @version
	`
	args := pgx.NamedArgs{
		"id":          product.id,
		"name":        product.name,
		"description": product.description,
		"category":    product.category,
		"price":       product.price.Amount,
		"stock":       product.stock,
		"reserved":    product.reserved,
		"is_active":   product.isActive,
		"version":     product.version,
	}

	row, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errs.ErrVersionConflict
	}
	product.version++
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	DELETE FROM products WHERE id = @id AND reserved = 0
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	row, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errors.New("can not delete")
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		SELECT id, name, description, category, price, stock, reserved, is_active, version
		FROM products
		WHERE id = @id
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	cmd := exec.QueryRow(ctx, query, args)
	var p Product
	if err := cmd.Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.category,
		&p.price.Amount,
		&p.stock,
		&p.reserved,
		&p.isActive,
		&p.version,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]Product, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		SELECT id, name, description, category, price, stock, reserved, is_active, version
		FROM products
		WHERE id = ANY(@ids)
		ORDER BY id
	`
	args := pgx.NamedArgs{
		"ids": ids,
	}
	rows, err := exec.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Product, 0, len(ids))

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.id,
			&p.name,
			&p.description,
			&p.category,
			&p.price.Amount,
			&p.stock,
			&p.reserved,
			&p.isActive,
			&p.version,
		); err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) != len(ids) {
		return nil, errors.New("some id not found")
	}

	return result, nil
}

func (r *Repository) List(ctx context.Context, filters ListFilters) ([]Product, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	SELECT id, name, description, category, price, stock, reserved, is_active, version
	FROM products
	WHERE 1=1
	`

	args := pgx.NamedArgs{}

	if filters.Category != nil {
		query += " AND category = @category"
		args["category"] = *filters.Category
	}
	if filters.MinPrice != nil {
		query += " AND price >= @min_price"
		args["min_price"] = *filters.MinPrice
	}
	if filters.MaxPrice != nil {
		query += " AND price <= @max_price"
		args["max_price"] = *filters.MaxPrice
	}
	if filters.IsActive != nil {
		query += " AND is_active = @is_active"
		args["is_active"] = *filters.IsActive
	}

	switch filters.SortBy {
	case SortByCreatedAt:
		query += " ORDER BY created_at"
	case SortByName:
		query += " ORDER BY name"
	case SortByPrice:
		query += " ORDER BY price"
	case SortByStock:
		query += " ORDER BY stock"
	default:
		query += " ORDER BY created_at"
	}
	if filters.SortDesc {
		query += " DESC"
	} else {
		query += " ASC"
	}

	query += " LIMIT @limit OFFSET @offset"
	args["limit"] = filters.Limit
	args["offset"] = filters.Offset

	rows, err := exec.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Product

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.id,
			&p.name,
			&p.description,
			&p.category,
			&p.price.Amount,
			&p.stock,
			&p.reserved,
			&p.isActive,
			&p.version,
		); err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
