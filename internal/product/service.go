package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	repo IProductRepository
}
type ServiceDeps struct {
	Repository IProductRepository
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo: deps.Repository,
	}
}

func (s *Service) CreateProduct(ctx context.Context, create DTOCreateProduct) (*Product, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, errors.New("id gen server error")
	}
	product, err := NewProduct(id, create.Name, create.Description, create.Category, create.Price, create.Stock, create.IsActive)
	if err != nil {
		return nil, err
	}
	return product, s.repo.Save(ctx, product)
}

func (s *Service) ChangeProduct(ctx context.Context, upd DTOUpdateProduct) (*Product, error) {
	product, err := s.repo.GetByID(ctx, upd.ID)
	if err != nil {
		return nil, err
	}
	if upd.Name != nil {
		if err := product.ChangeName(*upd.Name); err != nil {
			return nil, err
		}
	}
	if upd.Description != nil {
		if err := product.ChangeDescription(*upd.Description); err != nil {
			return nil, err
		}
	}
	if upd.Category != nil {
		if err := product.ChangeCategory(*upd.Category); err != nil {
			return nil, err
		}
	}
	if upd.Price != nil {
		if err := product.ChangePrice(*upd.Price); err != nil {
			return nil, err
		}
	}
	if upd.Stock != nil {
		if err := product.ChangeStock(*upd.Stock); err != nil {
			return nil, err
		}
	}
	if upd.IsActive != nil {
		product.ChangeIsActive(*upd.IsActive)
	}
	return product, s.repo.Save(ctx, product)
}

func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
