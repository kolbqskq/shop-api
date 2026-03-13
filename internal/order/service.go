package order

import (
	"context"
	"errors"
	"shop-api/internal/cart"
	"shop-api/internal/errs"
	"shop-api/internal/product"

	"github.com/google/uuid"
)

type Service struct {
	cartRepo          ICartRepository
	orderRepo         IOrderRepository
	productRepository IProductRepository
	txManager         ITxManager
	paymentService    IPaymentService
}

type ServiceDeps struct {
	CartRepository    ICartRepository
	OrderRepository   IOrderRepository
	ProductRepository IProductRepository
	TxManager         ITxManager
	PaymentService    IPaymentService
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		cartRepo:          deps.CartRepository,
		orderRepo:         deps.OrderRepository,
		productRepository: deps.ProductRepository,
		txManager:         deps.TxManager,
		paymentService:    deps.PaymentService,
	}
}

func (s *Service) CreateOrder(ctx context.Context, userID uuid.UUID) (*DTOOrder, error) {
	var order *Order

	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		cart, err := s.cartRepo.GetActiveCart(ctx, userID)
		if err != nil {
			return err
		}
		if len(cart.Items()) == 0 {
			return errs.ErrEmptyCart
		}
		ids := buildIds(cart)
		products, err := s.productRepository.GetByIDs(ctx, ids)
		if err != nil {
			return err
		}
		items, err := buildOrderItems(cart, products)
		if err != nil {
			return err
		}
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		order, err = NewOrder(id, userID, items)
		if err != nil {
			return err
		}
		productsMap := buildProductsMap(products)
		for _, v := range cart.Items() {
			product, ok := productsMap[v.ProductID]
			if !ok {
				return errs.ErrItemNotFound
			}
			if err := product.Reserve(v.Quantity); err != nil {
				return err
			}
			if err := s.productRepository.Save(ctx, &product); err != nil {
				return err
			}
		}
		if err := s.orderRepo.Create(ctx, order); err != nil {
			return err
		}
		cart.MarkAsOrdered()
		return s.cartRepo.Save(ctx, cart)
	})
	if err != nil {
		return nil, err
	}
	dto := buildDTOOrder(order)
	return dto, nil
}

func (s *Service) GetOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*DTOOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}
	return buildDTOOrder(order), nil
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID, l, o *int) ([]DTOOrder, error) {
	limit := 10
	if l != nil && *l > 0 && *l <= 20 {
		limit = *l
	}
	offset := 0
	if o != nil && *o >= 0 {
		offset = *o
	}
	orders, err := s.orderRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	ordersDTO := make([]DTOOrder, 0, len(orders))
	for _, v := range orders {
		ordersDTO = append(ordersDTO, *buildDTOOrder(v))
	}
	return ordersDTO, nil
}

func (s *Service) PayOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*DTOOrder, error) {
	var order *Order
	var err error

	err = s.txManager.WithTx(ctx, func(ctx context.Context) error {
		order, err = s.orderRepo.GetByID(ctx, orderID, userID)
		if err != nil {
			return err
		}
		if err := s.paymentService.Pay(ctx, orderID, order.Total().Amount); err != nil {
			return err
		}
		if err := order.Pay(); err != nil {
			return err
		}
		if err := s.orderRepo.Save(ctx, order); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return buildDTOOrder(order), nil
}

func buildDTOOrder(order *Order) *DTOOrder {
	return &DTOOrder{
		ID:        order.id,
		Status:    string(order.status),
		Total:     order.total.Amount,
		CreatedAt: order.createdAt,
		Items:     buildDTOOrderItems(order.items),
	}
}

func buildDTOOrderItems(items []OrderItem) []DTOOrderItem {
	dtoItems := make([]DTOOrderItem, 0, len(items))
	for _, v := range items {
		dtoItems = append(dtoItems, DTOOrderItem{
			ProductID: v.ProductID,
			Name:      v.Name,
			Quantity:  v.Quantity,
			Price:     v.Price.Amount,
		})
	}
	return dtoItems
}

func buildIds(cart *cart.Cart) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(cart.Items()))
	for _, v := range cart.Items() {
		ids = append(ids, v.ProductID)
	}
	return ids
}

func buildOrderItems(cart *cart.Cart, products []product.Product) ([]OrderItem, error) {
	res := make(map[uuid.UUID]product.Product, len(products))
	for _, v := range products {
		res[v.ID()] = v
	}
	orderItems := make([]OrderItem, 0, len(cart.Items()))
	for _, v := range cart.Items() {
		product, ok := res[v.ProductID]
		if !ok {
			return nil, errors.New("Товар не найден")
		}
		orderItems = append(orderItems, OrderItem{
			ProductID: v.ProductID,
			Name:      product.Name(),
			Quantity:  v.Quantity,
			Price:     product.Price(),
		})
	}
	return orderItems, nil
}

func buildProductsMap(products []product.Product) map[uuid.UUID]product.Product {
	productsMap := make(map[uuid.UUID]product.Product, len(products))
	for _, v := range products {
		productsMap[v.ID()] = v
	}
	return productsMap
}
