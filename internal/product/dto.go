package product

import (
	"github.com/google/uuid"
)

type DTOUpdateProduct struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Category    *string
	Price       *int64
	Stock       *int
	IsActive    *bool
}

type DTOCreateProduct struct {
	Name        string
	Description string
	Category    string
	Price       int64
	Stock       int
	IsActive    bool
}

type DTOListFilters struct {
	Limit    *int
	Offset   *int
	SortBy   *string
	SortDesc *bool

	Category *string
	MinPrice *int64
	MaxPrice *int64
	IsActive *bool
}

type DTOProduct struct {
	ID          uuid.UUID
	Name        string
	Description string
	Category    string
	Price       int64
	Available   int
}
