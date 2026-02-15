package order

import (
	"context"
	"errors"
	"shop-api/internal/product"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	cartRepo          ICartRepository
	orderRepo         IOrderRepository
	productRepository IProductRepository
}

type ServiceDeps struct {
	CartRepository    ICartRepository
	OrderRepository   IOrderRepository
	ProductRepository IProductRepository
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		cartRepo:          deps.CartRepository,
		orderRepo:         deps.OrderRepository,
		productRepository: deps.ProductRepository,
	}
}

func (s *Service) CreateFromCart(ctx context.Context, userID uuid.UUID) (*Order, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, errors.New("empty cart")
	}
	ids := make([]uuid.UUID, 0, len(cart.Items))
	for _, v := range cart.Items {
		ids = append(ids, v.ProductID)
	}
	products, err := s.productRepository.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	var items []OrderItem
	var reservations []product.Reservation
	for _, v := range cart.Items {
		prod, ok := products[v.ProductID]
		if !ok {
			return nil, errors.New("Товар не найден")
		}
		reservations = append(reservations, product.Reservation{
			ProductID: prod.ID,
			Quantity:  v.Quantity,
		})
		items = append(items, OrderItem{
			ProductID: prod.ID,
			Name:      prod.Name,
			Quantity:  v.Quantity,
			Price:     prod.Price,
		})
	}
	if err := s.productRepository.Reserve(ctx, reservations); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	order, err := NewOrder(id, userID, cart.ID, items, now)
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}
