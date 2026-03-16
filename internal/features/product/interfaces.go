package product

import (
	"context"
	"shop-api/internal/features/user"

	"github.com/google/uuid"
)

type IProductRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, product *Product) error
	Save(ctx context.Context, product *Product) error
	List(ctx context.Context, filters ListFilters) ([]Product, error)
}

type ITxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type IJWTService interface {
	ValidateAccessToken(tokenStr string) (uuid.UUID, string, error)
}

type IProductService interface {
	CreateProduct(ctx context.Context, create CreateProductRequest) (*DTOProduct, error)
	ChangeProduct(ctx context.Context, upd UpdateProductRequest) (*DTOProduct, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	GetList(ctx context.Context, filters ListFiltersRequest, role user.Role) ([]DTOProduct, error)
	GetByID(ctx context.Context, id uuid.UUID) (*DTOProduct, error)
}
