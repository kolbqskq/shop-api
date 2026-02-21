package order

import (
	"context"
	"shop-api/internal/cart"
	"shop-api/internal/product"

	"github.com/google/uuid"
)

type ICartRepository interface {
	GetByUserID(ctx context.Context, id uuid.UUID) (*cart.Cart, error)
	Save(ctx context.Context, cart *cart.Cart) error
	MakeAsOrdered(ctx context.Context, id uuid.UUID) error
}

type IOrderRepository interface {
	Save(ctx context.Context, order *Order) error
	MakeAsPaid(ctx context.Context, id uuid.UUID) error
	MakeAsCancelled(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetByUserID(ctx context.Context, id uuid.UUID, limit, offset int) ([]*Order, error)
}

type IProductRepository interface {
	Reserve(ctx context.Context, products []product.Reservation) ([]product.Product, error)
	Release(ctx context.Context, products []product.Reservation) error
	Commit(ctx context.Context, products []product.Reservation) error
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*product.Product, error)
}

type ITxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
