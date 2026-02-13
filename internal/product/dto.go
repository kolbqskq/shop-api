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
