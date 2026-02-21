package product

import (
	"context"

	"github.com/google/uuid"
)

type IProductRepository interface {
	Update(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, product *Product) error
	List(ctx context.Context, filters ListFilters) ([]Product, error)
}

type ITxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
