package cart

import (
	"shop-api/internal/money"

	"github.com/google/uuid"
)

type DTOCartItemView struct {
	ProductID uuid.UUID
	Name      string
	Price     money.Money
	IsActive  bool
	Quantity  int
}

type DTOCart struct {
	ID    uuid.UUID
	Items []DTOCartItemView
}

type DTOAddCartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type DTOUpdateCartItem struct {
	Quantity int       `json:"quantity"`
}