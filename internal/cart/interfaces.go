package cart

import (
	"context"
	"shop-api/internal/product"

	"github.com/google/uuid"
)

type ICardRepository interface {
	GetByUserID(ctx context.Context, id uuid.UUID) (*Cart, error)
	Save(ctx context.Context, cart *Cart) error
}

type IProductService interface {
	GetProductByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*product.Product, error)
}
