package order

import (
	"errors"
	"shop-api/internal/core/errs"
	"shop-api/internal/core/money"
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

func (o Order) Status() OrderStatus {
	return o.status
}

func (o Order) UserID() uuid.UUID {
	return o.userID
}

func (o Order) ID() uuid.UUID {
	return o.id
}

func (o Order) Items() []OrderItem {
	return o.items
}

func (o Order) Total() money.Money {
	return o.total
}

func (o *Order) Pay() error {
	if o.status != OrderStatusPending {
		return errs.ErrOrderNotPending
	}
	o.status = OrderStatusPaid
	return nil
}
