package order

import (
	"context"
	"errors"
	"shop-api/internal/cart"
	"shop-api/internal/product"
	"sort"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	cartRepo          ICartRepository
	orderRepo         IOrderRepository
	productRepository IProductRepository
	txManager         ITxManager
}

type ServiceDeps struct {
	CartRepository    ICartRepository
	OrderRepository   IOrderRepository
	ProductRepository IProductRepository
	TxManager         ITxManager
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		cartRepo:          deps.CartRepository,
		orderRepo:         deps.OrderRepository,
		productRepository: deps.ProductRepository,
		txManager:         deps.TxManager,
	}
}

func (s *Service) CreateFromCart(ctx context.Context, userID uuid.UUID) (*Order, error) {
	var res *Order

	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		cart, err := s.cartRepo.GetByUserID(ctx, userID) //Получаем корзину
		if err != nil {
			return err
		}
		if len(cart.Items) == 0 {
			return errors.New("empty cart")
		}
		reservations := buildReservationsFromCart(cart)
		products, err := s.productRepository.Reserve(ctx, reservations)
		if err != nil {
			return err
		}
		productsMap := buildProductsMap(products)

		orderItems, err := buildOrderItems(cart, productsMap)
		if err != nil {
			return err
		}
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		now := time.Now()
		order, err := NewOrder(id, userID, cart.ID, orderItems, now) // Создаем ордер
		if err != nil {
			return err
		}
		if err := s.orderRepo.Save(ctx, order); err != nil { // Сохраняем в бд
			return err
		}
		if err := s.cartRepo.MakeAsOrdered(ctx, cart.ID); err != nil { //Меняем статус корзины
			return err
		}
		res = order

		return nil
	})
	return res, err
}

func (s *Service) MarkAsPaid(ctx context.Context, orderID uuid.UUID) error {
	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {

		order, err := s.orderRepo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}

		if order.Status != OrderStatusPending {
			return errors.New("failed status order not pending")
		}

		reservations := buildReservationsFromOrder(order)

		if err := s.productRepository.Commit(ctx, reservations); err != nil {
			return err
		}

		if err := s.orderRepo.MakeAsPaid(ctx, orderID); err != nil {
			return err
		}
		return nil
	})

	return err
}

func (s *Service) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		order, err := s.orderRepo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}

		if order.Status == OrderStatusCancelled {
			return nil
		}

		if order.Status != OrderStatusPending {
			return errors.New("failed status order not pending")
		}

		reservations := buildReservationsFromOrder(order)

		if err := s.productRepository.Release(ctx, reservations); err != nil {
			return err
		}

		if err := s.orderRepo.MakeAsCancelled(ctx, orderID); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *Service) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*DTOOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return buildDTOOrder(order), nil
}

func (s *Service) GetOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]DTOOrder, error) {
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

func buildDTOOrder(order *Order) *DTOOrder {
	return &DTOOrder{
		ID:        order.ID,
		Status:    string(order.Status),
		Total:     order.Total.Amount,
		CreatedAt: order.CreatedAt,
		Items:     buildDTOOrderItems(order.Items),
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

func buildReservationsFromCart(cart *cart.Cart) []product.Reservation {
	reservations := make([]product.Reservation, 0, len(cart.Items))
	for _, v := range cart.Items {
		reservations = append(reservations, product.Reservation{
			ProductID: v.ProductID,
			Quantity:  v.Quantity,
		})
	}
	sort.Slice(reservations, func(i, j int) bool {
		return reservations[i].ProductID.String() < reservations[j].ProductID.String()
	})
	return reservations
}

func buildReservationsFromOrder(order *Order) []product.Reservation {
	reservations := make([]product.Reservation, 0, len(order.Items))
	for _, v := range order.Items {
		reservations = append(reservations, product.Reservation{
			ProductID: v.ProductID,
			Quantity:  v.Quantity,
		})
	}
	sort.Slice(reservations, func(i, j int) bool {
		return reservations[i].ProductID.String() < reservations[j].ProductID.String()
	})
	return reservations
}

func buildIds(cart *cart.Cart) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(cart.Items))
	for _, v := range cart.Items {
		ids = append(ids, v.ProductID)
	}
	return ids
}

func buildProductsMap(products []product.Product) map[uuid.UUID]product.Product {
	res := make(map[uuid.UUID]product.Product, len(products))
	for _, v := range products {
		res[v.ID] = v
	}
	return res
}

func buildOrderItems(cart *cart.Cart, products map[uuid.UUID]product.Product) ([]OrderItem, error) {
	orderItems := make([]OrderItem, 0, len(cart.Items))
	for _, v := range cart.Items {
		product, ok := products[v.ProductID]
		if !ok {
			return nil, errors.New("Товар не найден")
		}
		orderItems = append(orderItems, OrderItem{
			ProductID: v.ProductID,
			Name:      product.Name,
			Quantity:  v.Quantity,
			Price:     product.Price,
		})
	}
	return orderItems, nil
}
