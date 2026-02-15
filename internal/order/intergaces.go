package order

import (
	"context"
	"shop-api/internal/cart"
	"shop-api/internal/product"

	"github.com/google/uuid"
)

type ICartRepository interface {
	GetByUserID(ctx context.Context, id uuid.UUID) (*cart.Cart, error)
	MakeAsOrdered(ctx context.Context, id uuid.UUID) error
}

type IOrderRepository interface {
	Save(ctx context.Context, order *Order) error
}

type IProductRepository interface {
	Reserve(ctx context.Context, products []product.Reservation) error
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*product.Product, error)
}
