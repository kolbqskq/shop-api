package order

import (
	"context"
	"shop-api/internal/features/cart"
	"shop-api/internal/features/product"

	"github.com/google/uuid"
)

type ICartRepository interface {
	GetActiveCart(ctx context.Context, userID uuid.UUID) (*cart.Cart, error)
	Create(ctx context.Context, cart *cart.Cart) error
	Save(ctx context.Context, cart *cart.Cart) error
}

type IOrderRepository interface {
	Save(ctx context.Context, order *Order) error
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Order, error)
}

type IProductRepository interface {
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]product.Product, error)
	Save(ctx context.Context, product *product.Product) error
}

type ITxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type IOrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID) (*DTOOrder, error)
	GetOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*DTOOrder, error)
	ListByUser(ctx context.Context, userID uuid.UUID, l, o *int) ([]DTOOrder, error)
	PayOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*DTOOrder, error)
}

type IJWTService interface {
	ValidateAccessToken(tokenStr string) (uuid.UUID, string, error)
}

type IPaymentService interface {
	Pay(ctx context.Context, orderID uuid.UUID, amount int64) error
}
