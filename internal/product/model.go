package product

import (
	"errors"
	"shop-api/pkg/money"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Category    string
	Price       money.Money
	Stock       int
	Reserved    int
	IsActive    bool
	Version     int64
}

func NewProduct(id uuid.UUID, name, description, category string, price int64, stock int, isActive bool) (*Product, error) {
	product := &Product{
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		Price:       money.Money{Amount: price},
		Stock:       stock,
		IsActive:    isActive,
		Reserved:    0,
		Version:     0,
	}
	if err := product.Validate(); err != nil {
		return nil, err
	}
	return product, nil
}

func (p *Product) Validate() error {
	if p.Name == "" {
		return errors.New("empty name")
	}
	if p.Price.Amount < 0 {
		return errors.New("invalid price")
	}
	if p.Stock < p.Reserved {
		return errors.New("invalid stock")
	}
	return nil
}

func (p *Product) ChangeName(name string) error {
	if name == "" {
		return errors.New("empty name")
	}
	p.Name = name
	return nil
}

func (p *Product) ChangeDescription(description string) error {
	if description == "" {
		return errors.New("empty desc")
	}
	p.Description = description
	return nil
}

func (p *Product) ChangeCategory(category string) error {
	if category == "" {
		return errors.New("empty category")
	}
	p.Category = category
	return nil
}

func (p *Product) ChangePrice(price int64) error {
	if price < 0 {
		return errors.New("invalid price")
	}
	p.Price.Amount = price
	return nil
}

func (p *Product) ChangeStock(stock int) error {
	if stock < p.Reserved {
		return errors.New("invalid stock")
	}
	p.Stock = stock
	return nil
}

func (p *Product) ChangeIsActive(isActive bool) {
	p.IsActive = isActive
}

func (p *Product) Reserve(amount int) error {
	if amount <= 0 {
		return errors.New("invalid reserve amount")
	}
	if !p.IsActive {
		return errors.New("product inactive")
	}
	if p.Reserved+amount > p.Stock {
		return errors.New("cant be reserved")
	}
	p.Reserved += amount
	return nil
}

func (p *Product) Commit(amount int) error {
	if p.Reserved < amount || p.Stock < amount {
		return errors.New("invalid release")
	}
	p.Reserved -= amount
	p.Stock -= amount
	return nil
}

func (p *Product) Unreserve(amount int) error {
	if p.Reserved < amount {
		return errors.New("invalid unreserve")
	}
	p.Reserved -= amount
	return nil
}
