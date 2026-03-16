package cart

import (
	"shop-api/internal/core/errs"

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
}

type Cart struct {
	id      uuid.UUID
	userID  uuid.UUID
	status  CartStatus
	items   []CartItem
	version int64
}

func NewCart(id uuid.UUID, userID uuid.UUID) *Cart {
	return &Cart{
		id:      id,
		userID:  userID,
		status:  CartStatusActive,
		items:   []CartItem{},
		version: 0,
	}
}

func (c *Cart) AddItem(productID uuid.UUID, qty int) error {
	if c.status != CartStatusActive {
		return errs.ErrCartNotActive
	}
	if qty <= 0 {
		return errs.ErrInvalidQuantity
	}
	for k, v := range c.items {
		if v.ProductID == productID {
			c.items[k].Quantity += qty
			return nil
		}
	}
	c.items = append(c.items, CartItem{
		ProductID: productID,
		Quantity:  qty,
	})
	return nil
}

func (c *Cart) ChangeQuantityItem(productID uuid.UUID, qty int) error {
	if c.status != CartStatusActive {
		return errs.ErrCartNotActive
	}
	if qty < 0 {
		return errs.ErrInvalidQuantity
	}
	for k := range c.items {
		if c.items[k].ProductID == productID {
			if qty == 0 {
				c.removeByIndex(k)
				return nil
			}

			c.items[k].Quantity = qty
			return nil
		}
	}
	return errs.ErrProductNotFound
}

func (c *Cart) RemoveItem(productID uuid.UUID) error {
	if c.status != CartStatusActive {
		return errs.ErrCartNotActive
	}
	for k := range c.items {
		if c.items[k].ProductID == productID {
			c.removeByIndex(k)
			return nil
		}
	}
	return errs.ErrProductNotFound
}

func (c *Cart) MarkAsExpired() error {
	if c.status != CartStatusActive {
		return errs.ErrCartNotActive
	}

	c.status = CartStatusExpired

	return nil
}

func (c *Cart) ClearItems() error {
	if c.status != CartStatusActive {
		return errs.ErrCartNotActive
	}

	c.items = []CartItem{}

	return nil
}

func (c *Cart) MarkAsOrdered() error {
	if c.status != CartStatusActive {
		return errs.ErrCartNotActive
	}

	c.status = CartStatusOrdered

	return nil
}

func (c *Cart) ID() uuid.UUID {
	return c.id
}

func (c *Cart) Items() []CartItem {
	return c.items
}

func (c *Cart) Status() CartStatus {
	return c.status
}

func (c *Cart) removeByIndex(i int) {
	c.items[i] = c.items[len(c.items)-1]
	c.items = c.items[:len(c.items)-1]
}
