package cart

import (
	"context"
	"errors"
	"shop-api/internal/errs"
	"shop-api/internal/product"

	"github.com/google/uuid"
)

type Service struct {
	repo        ICardRepository
	productRepo IProductRepository
	txManager   ITxManager
}
type ServiceDeps struct {
	Repository        ICardRepository
	ProductRepository IProductRepository
	TxManager         ITxManager
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo:        deps.Repository,
		productRepo: deps.ProductRepository,
		txManager:   deps.TxManager,
	}
}

func (s *Service) getOrCreateActiveCart(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	cart, err := s.repo.GetActiveCart(ctx, userID)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, errs.ErrCartNotFound) {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	cart = NewCart(id, userID)
	if err := s.repo.Create(ctx, cart); err != nil {
		if errors.Is(err, errs.ErrCartAlreadyExists) {
			return s.repo.GetActiveCart(ctx, userID)
		}
		return nil, err
	}

	return cart, nil
}

func (s *Service) AddToCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (*DTOCart, error) {
	if _, err := s.productRepo.GetByID(ctx, productID); err != nil {
		return nil, err
	}

	var cart *Cart

	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.productRepo.GetByID(ctx, productID); err != nil {
			return err
		}
		c, err := s.getOrCreateActiveCart(ctx, userID)
		if err != nil {
			return err
		}
		if err := c.AddItem(productID, qty); err != nil {
			return err
		}
		cart = c
		if err := s.repo.Save(ctx, cart); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.buildDTOCart(ctx, cart)
}

func (s *Service) UpdateFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (*DTOCart, error) {

	var cart *Cart

	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		c, err := s.repo.GetActiveCart(ctx, userID)
		if err != nil {
			return err
		}
		if err := c.ChangeQuantityItem(productID, qty); err != nil {
			return err
		}
		cart = c
		if err := s.repo.Save(ctx, c); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.buildDTOCart(ctx, cart)
}

func (s *Service) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*DTOCart, error) {

	var cart *Cart

	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		c, err := s.repo.GetActiveCart(ctx, userID)
		if err != nil {
			return err
		}
		if err := c.RemoveItem(productID); err != nil {
			return err
		}
		cart = c
		if err := s.repo.Save(ctx, c); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.buildDTOCart(ctx, cart)
}

func (s *Service) GetActiveCart(ctx context.Context, userID uuid.UUID) (*DTOCart, error) {
	cart, err := s.getOrCreateActiveCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.buildDTOCart(ctx, cart)
}

func (s *Service) ClearCart(ctx context.Context, userID uuid.UUID) (*DTOCart, error) {
	var cart *Cart
	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		c, err := s.repo.GetActiveCart(ctx, userID)
		if err != nil {
			return err
		}
		c.ClearItems()

		if err := s.repo.Save(ctx, c); err != nil {
			return err
		}
		cart = c
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.buildDTOCart(ctx, cart)
}

func (s *Service) buildDTOCart(ctx context.Context, cart *Cart) (*DTOCart, error) {
	if len(cart.items) == 0 {
		return &DTOCart{
			ID:    cart.id,
			Items: []DTOCartItemView{},
		}, nil
	}
	ids := make([]uuid.UUID, 0, len(cart.items))
	for _, v := range cart.items {
		ids = append(ids, v.ProductID)
	}
	products, err := s.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	productMap := make(map[uuid.UUID]product.Product, len(products))
	for _, v := range products {
		productMap[v.ID()] = v
	}
	dto := &DTOCart{
		ID:    cart.id,
		Items: make([]DTOCartItemView, 0, len(cart.items)),
	}

	for _, v := range cart.items {
		p, ok := productMap[v.ProductID]
		if !ok {
			continue
		}
		dto.Items = append(dto.Items, DTOCartItemView{
			ProductID: p.ID(),
			Name:      p.Name(),
			Price:     p.Price(),
			IsActive:  p.IsActive(),
			Quantity:  v.Quantity,
		})
	}
	return dto, nil
}
