package product

import (
	"shop-api/internal/money"

	"github.com/google/uuid"
)

type CreateProductRequest struct {
	Name        string
	Description string
	Category    string
	Price       money.Money
	Stock       int
	IsActive    bool
}

type UpdateProductRequest struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Category    *string
	Price       *money.Money
	Stock       *int
	IsActive    *bool
}

type ListFiltersRequest struct {
	Limit    *int
	Offset   *int
	SortBy   *ProductSortField
	SortDesc *bool

	Category *string
	MinPrice *int64
	MaxPrice *int64
	IsActive *bool
}
