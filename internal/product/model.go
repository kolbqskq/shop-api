package product

import (
	"errors"
	"shop-api/internal/errs"
	"shop-api/internal/money"

	"github.com/google/uuid"
)

type Product struct {
	id          uuid.UUID
	name        string
	description string
	category    string
	price       money.Money
	stock       int
	reserved    int
	isActive    bool
	version     int64
}

type ProductSortField string

const (
	SortByCreatedAt ProductSortField = "created_at"
	SortByPrice     ProductSortField = "price"
	SortByName      ProductSortField = "name"
	SortByStock     ProductSortField = "stock"
)

type ListFilters struct {
	Limit    int
	Offset   int
	SortBy   ProductSortField
	SortDesc bool

	Category *string
	MinPrice *int64
	MaxPrice *int64
	IsActive *bool
}

func NewProduct(id uuid.UUID, name, description, category string, price money.Money, stock int, isActive bool) (*Product, error) {
	product := &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		price:       price,
		stock:       stock,
		isActive:    isActive,
		reserved:    0,
		version:     0,
	}
	if err := product.Validate(); err != nil {
		return nil, err
	}
	return product, nil
}

func (p *Product) Validate() error {
	if p.name == "" {
		return errors.New("empty name")
	}
	if p.price.Amount < 0 {
		return errs.ErrInvalidPrice
	}
	if p.stock < p.reserved {
		return errs.ErrInvalidStock
	}
	return nil
}

func (p *Product) ChangeName(name string) error {
	if name == "" {
		return errors.New("empty name")
	}
	p.name = name
	return nil
}

func (p *Product) ChangeDescription(description string) error {
	if description == "" {
		return errors.New("empty desc")
	}
	p.description = description
	return nil
}

func (p *Product) ChangeCategory(category string) error {
	if category == "" {
		return errors.New("empty category")
	}
	p.category = category
	return nil
}

func (p *Product) ChangePrice(price money.Money) error {
	if price.Amount < 0 {
		return errors.New("invalid price")
	}
	p.price = price
	return nil
}

func (p *Product) ChangeStock(qty int) error {
	if qty < p.reserved {
		return errors.New("invalid stock")
	}
	p.stock = qty
	return nil
}

func (p *Product) ChangeIsActive(isActive bool) {
	p.isActive = isActive
}

func (p *Product) Reserve(qty int) error {
	if qty > p.stock-p.reserved {
		return errs.ErrNotEnoughStock
	}
	p.reserved += qty
	return nil
}

func (p *Product) ID() uuid.UUID {
	return p.id
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Description() string {
	return p.description
}

func (p *Product) Category() string {
	return p.category
}

func (p *Product) Price() money.Money {
	return p.price
}

func (p *Product) Stock() int {
	return p.stock
}

func (p *Product) Reserved() int {
	return p.reserved
}

func (p *Product) IsActive() bool {
	return p.isActive
}

func (p *Product) Available() int {
	return p.stock - p.reserved
}
