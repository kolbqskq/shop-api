package cart

import (
	"context"
	"shop-api/internal/product"

	"github.com/google/uuid"
)

type ICardRepository interface {
	GetActiveCart(ctx context.Context, userID uuid.UUID) (*Cart, error)
	Create(ctx context.Context, cart *Cart) error
	Save(ctx context.Context, cart *Cart) error
}

type IProductRepository interface {
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]product.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error)
}

type ITxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type ICartService interface {
	AddToCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (*DTOCart, error)
	UpdateFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (*DTOCart, error)
	RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*DTOCart, error)
	GetActiveCart(ctx context.Context, userID uuid.UUID) (*DTOCart, error)
	ClearCart(ctx context.Context, userID uuid.UUID) (*DTOCart, error)
}

type IJWTService interface {
	ValidateAccessToken(tokenStr string) (uuid.UUID, string, error)
}
