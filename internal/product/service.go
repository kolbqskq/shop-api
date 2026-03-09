package product

import (
	"context"
	"shop-api/internal/errs"

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

func (s *Service) CreateProduct(ctx context.Context, create CreateProductRequest) (*DTOProduct, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	product, err := NewProduct(id, create.Name, create.Description, create.Category, create.Price, create.Stock, create.IsActive)
	if err != nil {
		return nil, err
	}
	return buildDTOProduct(product), s.repo.Create(ctx, product)
}

func (s *Service) ChangeProduct(ctx context.Context, upd UpdateProductRequest) (*DTOProduct, error) {
	if upd.ID == uuid.Nil {
		return nil, errs.ErrMissingID
	}

	if upd.Name == nil && upd.Description == nil && upd.Category == nil && upd.Price == nil && upd.Stock == nil && upd.IsActive == nil {
		return nil, errs.ErrNothingToUpdate
	}

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
	return buildDTOProduct(product), s.repo.Save(ctx, product)
}

func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetList(ctx context.Context, filters ListFiltersRequest) ([]DTOProduct, error) {
	limit := 20
	if filters.Limit != nil && *filters.Limit > 0 && *filters.Limit <= 100 {
		limit = *filters.Limit
	}
	offset := 0
	if filters.Offset != nil && *filters.Offset >= 0 {
		offset = *filters.Offset
	}
	sortBy := SortByCreatedAt
	if filters.SortBy != nil {
		sortBy = *filters.SortBy
	}
	sortDesc := false
	if filters.SortDesc != nil {
		sortDesc = *filters.SortDesc
	}
	filter := ListFilters{
		Limit:    limit,
		Offset:   offset,
		SortBy:   sortBy,
		SortDesc: sortDesc,

		Category: filters.Category,
		MinPrice: filters.MinPrice,
		MaxPrice: filters.MaxPrice,
		IsActive: filters.IsActive,
	}
	products, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return buildDTOProductSlice(products), nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*DTOProduct, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return buildDTOProduct(product), nil
}

func buildDTOProduct(product *Product) *DTOProduct {
	return &DTOProduct{
		ID:          product.id.String(),
		Name:        product.name,
		Description: product.description,
		Category:    product.category,
		Price:       product.price.Amount,
		Available:   product.Available(),
	}
}

func buildDTOProductSlice(products []Product) []DTOProduct {
	dto := make([]DTOProduct, 0, len(products))
	for _, v := range products {
		dto = append(dto, *buildDTOProduct(&v))
	}
	return dto
}
