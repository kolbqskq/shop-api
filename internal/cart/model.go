package cart

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type CartStatus string

const (
	CartStatusActive  CartStatus = "active"
	CartStatusExpired CartStatus = "expired"
	CartStatusOrdered CartStatus = "ordered"
)

type CartItem struct {
	ProductID uuid.UUID
	Quantity  int
	AddedAt   time.Time
}

type Cart struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Status        CartStatus
	LastUpdatedAt time.Time
	Items         []CartItem
	Version       int64
}

func NewCart(id uuid.UUID, userID uuid.UUID, now time.Time) *Cart {
	return &Cart{
		ID:            id,
		UserID:        userID,
		Status:        CartStatusActive,
		LastUpdatedAt: now,
		Items:         []CartItem{},
		Version:       0,
	}
}

func (c *Cart) AddItem(productID uuid.UUID, qty int, now time.Time) error {
	if c.Status != CartStatusActive {
		return errors.New("cart inactive")
	}
	if qty <= 0 {
		return errors.New("qty should be > 0")
	}
	for k, v := range c.Items {
		if v.ProductID == productID {
			c.Items[k].Quantity += qty
			return nil
		}
	}
	c.Items = append(c.Items, CartItem{
		ProductID: productID,
		Quantity:  qty,
		AddedAt:   time.Now(),
	})
	c.LastUpdatedAt = now
	return nil
}

func (c *Cart) RemoveItem(productID uuid.UUID, now time.Time) error {
	if c.Status != CartStatusActive {
		return errors.New("cart inactive")
	}
	c.LastUpdatedAt = now
	for k, v := range c.Items {
		if v.ProductID == productID {
			c.Items = append(c.Items[:k], c.Items[k+1:]...)
			return nil
		}
	}
	return errors.New("failed remove product not found")
}

func (c *Cart) ChangeQuantityItem(productID uuid.UUID, qty int, now time.Time) error {
	if c.Status != CartStatusActive {
		return errors.New("cart inactive")
	}
	if qty <= 0 {
		return errors.New("qty should be > 0")
	}
	c.LastUpdatedAt = now
	for k, v := range c.Items {
		if v.ProductID == productID {
			c.Items[k].Quantity += qty
			return nil
		}
	}
	return errors.New("failed change qty product not found")
}

func (c *Cart) CartDoExpired() {
	c.Status = CartStatusExpired
}

func (c *Cart) ClearItems(now time.Time) {
	c.Items = []CartItem{}
	c.LastUpdatedAt = now
}
