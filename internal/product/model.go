package product

import (
	"errors"
	"shop-api/internal/errs"
	"shop-api/internal/money"
	"time"

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
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Reservation struct {
	ProductID uuid.UUID
	Quantity  int
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
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		Price:       price,
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
		return errs.ErrInvalidPrice
	}
	if p.Stock < p.Reserved {
		return errs.ErrInvalidStock
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

func (p *Product) ChangePrice(price money.Money) error {
	if price.Amount < 0 {
		return errors.New("invalid price")
	}
	p.Price = price
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
