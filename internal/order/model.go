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
	id        uuid.UUID
	userID    uuid.UUID
	status    OrderStatus
	items     []OrderItem
	total     money.Money
	createdAt time.Time
	version   int64
}

func NewOrder(id, userID uuid.UUID, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("cart is empty")
	}
	return &Order{
		id:      id,
		userID:  userID,
		status:  OrderStatusPending,
		items:   items,
		total:   calculateTotal(items),
		version: 0,
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
