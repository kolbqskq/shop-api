package cart

import (
	"context"
	"errors"
	"shop-api/pkg/errs"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo        ICardRepository
	productRepo IProductRepository
}
type ServiceDeps struct {
	Repository        ICardRepository
	ProductRepository IProductRepository
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo:        deps.Repository,
		productRepo: deps.ProductRepository,
	}
}

func (s *Service) getOrCreateActiveCart(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	cart, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			id, err := uuid.NewV7()
			if err != nil {
				return nil, err
			}
			now := time.Now()
			cart = NewCart(id, userID, now)
			if err := s.repo.Save(ctx, cart); err != nil {
				if errors.Is(err, errs.ErrCartAlreadyExist) {
					return s.repo.GetByID(ctx, userID)
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return cart, nil
}

func (s *Service) buildDTOCart(ctx context.Context, cart *Cart) (*DTOCart, error) {
	if len(cart.Items) == 0 {
		return &DTOCart{
			ID:    cart.ID,
			Items: []DTOCartItemView{},
		}, nil
	}
	ids := make([]uuid.UUID, 0, len(cart.Items))
	for _, v := range cart.Items {
		ids = append(ids, v.ProductID)
	}
	products, err := s.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	dto := &DTOCart{
		ID:    cart.ID,
		Items: make([]DTOCartItemView, 0, len(cart.Items)),
	}
	for _, v := range cart.Items {
		product, ok := products[v.ProductID]
		if !ok {
			continue
		}
		dto.Items = append(dto.Items, DTOCartItemView{
			ProductID: product.ID,
			Name:      product.Name,
			Price:     product.Price,
			IsActive:  product.IsActive,
			Quantity:  v.Quantity,
		})
	}
	return dto, nil
}

func (s *Service) AddToCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) error {
	cart, err := s.getOrCreateActiveCart(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	if err := cart.AddItem(productID, qty, now); err != nil {
		return err
	}

	return s.repo.Save(ctx, cart)
}

func (s *Service) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	cart, err := s.getOrCreateActiveCart(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	if err := cart.RemoveItem(productID, now); err != nil {
		return err
	}

	return s.repo.Save(ctx, cart)
}

func (s *Service) ChangeQuantityFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) error {
	cart, err := s.getOrCreateActiveCart(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	if err := cart.ChangeQuantityItem(productID, qty, now); err != nil {
		return err
	}
	return s.repo.Save(ctx, cart)
}

func (s *Service) GetActiveCart(ctx context.Context, userID uuid.UUID) (*DTOCart, error) {
	cart, err := s.getOrCreateActiveCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.buildDTOCart(ctx, cart)
}

func (s *Service) ClearCart(ctx context.Context, userID uuid.UUID) (*DTOCart, error) {
	cart, err := s.getOrCreateActiveCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	cart.ClearItems(now)

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return s.buildDTOCart(ctx, cart)
}
