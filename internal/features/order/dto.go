package order

import (
	"time"

	"github.com/google/uuid"
)

type DTOOrder struct {
	ID        uuid.UUID
	Status    string
	Total     int64
	CreatedAt time.Time
	Items     []DTOOrderItem
}

type DTOOrderItem struct {
	ProductID uuid.UUID
	Name      string
	Quantity  int
	Price     int64
}

type DTOOrdersList struct {
	Limit  *int `form:"limit"`
	Offset *int `form:"offset"`
}
