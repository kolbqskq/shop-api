package order

import (
	"errors"
	"shop-api/internal/money"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID uuid.UUID
	Name      string
	Quantity  int
	Price     money.Money
}

type Order struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CartID    uuid.UUID
	Status    OrderStatus
	Items     []OrderItem
	Total     money.Money
	CreatedAt time.Time
	Version   int64
}

type Reservation struct {
	ProductID uuid.UUID
	Quantity  int
}

func NewOrder(id, userID, cartID uuid.UUID, items []OrderItem, now time.Time) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("cart is empty")
	}
	return &Order{
		ID:        id,
		UserID:    userID,
		CartID:    cartID,
		Status:    OrderStatusPending,
		Items:     items,
		Total:     calculateTotal(items),
		CreatedAt: now,
		Version:   0,
	}, nil
}

func calculateTotal(items []OrderItem) money.Money {
	var total int64
	for _, v := range items {
		total += v.Price.Amount * int64(v.Quantity)
	}
	return money.Money{
		Amount: total,
	}
}
