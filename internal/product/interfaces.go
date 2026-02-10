package product

import (
	"context"

	"github.com/google/uuid"
)

type IProductRepository interface {
	Save(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
