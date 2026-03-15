package order

import (
	"context"
	"errors"
	"shop-api/internal/database"
	"shop-api/internal/errs"
	"shop-api/internal/money"
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

func (r *Repository) Create(ctx context.Context, order *Order) error {
	exec := database.Executor(ctx, r.dbPool)

	tx, ok := exec.(pgx.Tx)
	if !ok {
		return errors.New("Create must be called inside transaction")
	}

	query :=
		`
	INSERT INTO orders (id, user_id, total)
	VALUES (@id, @user_id, @total)
	`
	args := pgx.NamedArgs{
		"id":      order.id,
		"user_id": order.userID,
		"total":   order.total.Amount,
	}
	_, err := tx.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	batch := &pgx.Batch{}
	for _, v := range order.items {
		batch.Queue(`
			INSERT INTO order_items (order_id, product_id, name, quantity, price)
			VALUES ($1, $2, $3, $4, $5)
		`, order.id, v.ProductID, v.Name, v.Quantity, v.Price.Amount)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range order.items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Order, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
		SELECT id, user_id, status, total, created_at, version
		FROM orders
		WHERE id = @id AND user_id = @user_id
	`
	args := pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	}

	row := exec.QueryRow(ctx, query, args)
	order := &Order{}
	if err := row.Scan(&order.id, &order.userID, &order.status, &order.total.Amount, &order.createdAt, &order.version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrOrderNotFound
		}
		return nil, err
	}

	query =
		`
		SELECT product_id, name, quantity, price
		FROM order_items
		WHERE order_id = @id
		ORDER BY product_id
	`
	rows, err := exec.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ProductID, &item.Name, &item.Quantity, &item.Price.Amount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	order.items = items
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return order, nil
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Order, error) {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	SELECT
		o.id,
		o.status,
		o.total,
		o.created_at,
		o.version,
		oi.product_id,
		oi.name,
		oi.quantity,
		oi.price
	FROM (
		SELECT id, status, total, created_at, version
		FROM orders
		WHERE user_id = @user_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	) o
	LEFT JOIN order_items oi ON oi.order_id = o.id
	ORDER BY o.created_at DESC
	`
	args := pgx.NamedArgs{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}
	rows, err := exec.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ordersMap := make(map[uuid.UUID]*Order, limit)
	res := make([]*Order, 0, limit)

	for rows.Next() {
		var (
			orderID   uuid.UUID
			status    OrderStatus
			total     int64
			createdAt time.Time
			version   int64
			productID *uuid.UUID
			name      *string
			quantity  *int
			price     *int64
		)

		if err := rows.Scan(&orderID, &status, &total, &createdAt, &version, &productID, &name, &quantity, &price); err != nil {
			return nil, err
		}
		order, ok := ordersMap[orderID]
		if !ok {
			order = &Order{
				id:        orderID,
				userID:    userID,
				status:    status,
				total:     money.Money{Amount: total},
				createdAt: createdAt,
				version:   version,
			}
			ordersMap[orderID] = order
			res = append(res, order)
		}
		if productID != nil {
			order.items = append(order.items, OrderItem{
				ProductID: *productID,
				Name:      *name,
				Quantity:  *quantity,
				Price:     money.Money{Amount: *price},
			})
		}

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Repository) Save(ctx context.Context, order *Order) error {
	exec := database.Executor(ctx, r.dbPool)

	query :=
		`
	UPDATE orders
	SET 
		status = @status,
		version = version + 1
	WHERE id = @id AND version = @version
	`
	args := pgx.NamedArgs{
		"id":      order.id,
		"status":  order.status,
		"version": order.version,
	}

	row, err := exec.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errs.ErrVersionConflict
	}
	return nil
}
