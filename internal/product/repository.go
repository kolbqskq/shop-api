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

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	DELETE FROM products WHERE id = @id AND reserved = 0
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	cmd, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("can not delete")
	}
	return nil
}

func (r *Repository) Reserve(ctx context.Context, products []Reservation) ([]Product, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	UPDATE products
	SET
		reserved = reserved + @quantity
	WHERE id = @id AND is_active = TRUE AND stock - reserved >= @quantity
	RETURNING id, name, description, category, price, stock, reserved, is_active, version, created_at, updated_at
	`
	var result []Product

	for _, v := range products {

		if v.Quantity <= 0 {
			return nil, errors.New("invalid quantity")
		}

		args := pgx.NamedArgs{
			"id":       v.ProductID,
			"quantity": v.Quantity,
		}

		var p Product

		err := exec.QueryRow(ctx, query, args).Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.Price.Amount,
			&p.Stock,
			&p.Reserved,
			&p.IsActive,
			&p.Version,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, errs.ErrNotEnoughStock
			}
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil

}

func (r *Repository) Commit(ctx context.Context, products []Reservation) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	UPDATE products
	SET
		reserved = reserved - @quantity,
		stock = stock - @quantity
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
			return errs.ErrNotEnoughStock
		}
	}
	return nil
}

func (r *Repository) Release(ctx context.Context, products []Reservation) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	UPDATE products
	SET
		reserved = reserved - @quantity
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

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		SELECT id, name, description, category, price, stock, reserved, is_active, version, created_at, updated_at
		FROM products
		WHERE id = @id
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	cmd := exec.QueryRow(ctx, query, args)
	var p Product
	if err := cmd.Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Category,
		&p.Price.Amount,
		&p.Stock,
		&p.Reserved,
		&p.IsActive,
		&p.Version,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]Product, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		SELECT id, name, description, category, price, stock, reserved, is_active, version, created_at, updated_at
		FROM products
		WHERE id = ANY(@ids)
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
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.Price.Amount,
			&p.Stock,
			&p.Reserved,
			&p.IsActive,
			&p.Version,
			&p.CreatedAt,
			&p.UpdatedAt,
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
	SELECT id, name, description, category, price, stock, reserved, is_active, version, created_at, updated_at
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
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.Price.Amount,
			&p.Stock,
			&p.Reserved,
			&p.IsActive,
			&p.Version,
			&p.CreatedAt,
			&p.UpdatedAt,
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
