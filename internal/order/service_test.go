package order_test

import (
	"context"
	"shop-api/internal/cart"
	"shop-api/internal/errs"
	"shop-api/internal/money"
	"shop-api/internal/order"
	"shop-api/internal/product"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type MockTx struct {
}

func (m *MockTx) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type MockCartRepository struct {
	GetCalled    bool
	SaveCalled   bool
	CartToReturn *cart.Cart
	CartSaved    *cart.Cart
	GetErr       error
}

func (m *MockCartRepository) GetActiveCart(ctx context.Context, userID uuid.UUID) (*cart.Cart, error) {
	m.GetCalled = true
	return m.CartToReturn, m.GetErr
}
func (m *MockCartRepository) Create(ctx context.Context, cart *cart.Cart) error {
	return nil
}
func (m *MockCartRepository) Save(ctx context.Context, cart *cart.Cart) error {
	m.SaveCalled = true
	m.CartSaved = cart
	return nil
}

type MockOrderRepository struct {
	CreateCalled bool
	SaveCalled   bool
	OrderCreated *order.Order
	OrderSaved   *order.Order
}

func (m *MockOrderRepository) Save(ctx context.Context, order *order.Order) error {
	m.SaveCalled = true
	m.OrderSaved = order
	return nil
}
func (m *MockOrderRepository) Create(ctx context.Context, order *order.Order) error {
	m.CreateCalled = true
	m.OrderCreated = order
	return nil
}
func (m *MockOrderRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*order.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*order.Order, error) {
	return nil, nil
}

type MockProductRepository struct {
	GetByIDsCalled   bool
	SaveCalled       bool
	ProductsToReturn []product.Product
	ProductsSaved    []product.Product
}

func (m *MockProductRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]product.Product, error) {
	m.GetByIDsCalled = true
	return m.ProductsToReturn, nil
}
func (m *MockProductRepository) Save(ctx context.Context, product *product.Product) error {
	m.SaveCalled = true
	m.ProductsSaved = append(m.ProductsSaved, *product)
	return nil
}

func TestCreateFromCart_CartExist_CreateOrder(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	cartID, err := uuid.NewV7()
	require.NoError(t, err)
	productID, err := uuid.NewV7()
	require.NoError(t, err)
	c := cart.NewCart(cartID, userID)
	c.AddItem(productID, 7)
	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	tx := &MockTx{}
	cartRepo := &MockCartRepository{
		CartToReturn: c,
	}
	orderRepo := &MockOrderRepository{}
	productRepo := &MockProductRepository{
		ProductsToReturn: []product.Product{*p},
	}

	service := order.NewService(order.ServiceDeps{
		CartRepository:    cartRepo,
		OrderRepository:   orderRepo,
		ProductRepository: productRepo,
		TxManager:         tx,
	})

	ctx := context.Background()
	dtoOrder, err := service.CreateFromCart(ctx, userID)
	require.NoError(t, err)

	require.NotNil(t, dtoOrder)

	require.True(t, cartRepo.GetCalled)
	require.True(t, cartRepo.SaveCalled)
	require.True(t, orderRepo.CreateCalled)
	require.False(t, orderRepo.SaveCalled)
	require.True(t, productRepo.SaveCalled)
	require.True(t, productRepo.GetByIDsCalled)

	require.Len(t, productRepo.ProductsSaved, 1)

	require.Equal(t, 7, productRepo.ProductsSaved[0].Reserved())
	require.Equal(t, 3, productRepo.ProductsSaved[0].Available())
	require.Equal(t, cart.CartStatusOrdered, cartRepo.CartSaved.Status())

	require.Len(t, orderRepo.OrderCreated.Items(), 1)

	require.Equal(t, order.OrderStatusPending, orderRepo.OrderCreated.Status())
	require.Equal(t, productID, orderRepo.OrderCreated.Items()[0].ProductID)
	require.Equal(t, 7, orderRepo.OrderCreated.Items()[0].Quantity)
	require.Equal(t, int64(100), orderRepo.OrderCreated.Items()[0].Price.Amount)

	require.Len(t, dtoOrder.Items, 1)

	require.Equal(t, orderRepo.OrderCreated.ID(), dtoOrder.ID)
	require.Equal(t, 7, dtoOrder.Items[0].Quantity)
	require.Equal(t, int64(700), dtoOrder.Total)
}

func TestCreateFromCart_CartEmpty_ErrEmptyCart(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	cartID, err := uuid.NewV7()
	require.NoError(t, err)
	c := cart.NewCart(cartID, userID)
	tx := &MockTx{}
	cartRepo := &MockCartRepository{
		CartToReturn: c,
	}
	orderRepo := &MockOrderRepository{}
	productRepo := &MockProductRepository{}

	service := order.NewService(order.ServiceDeps{
		CartRepository:    cartRepo,
		OrderRepository:   orderRepo,
		ProductRepository: productRepo,
		TxManager:         tx,
	})

	ctx := context.Background()
	dtoOrder, err := service.CreateFromCart(ctx, userID)
	require.ErrorIs(t, err, errs.ErrEmptyCart)

	require.Nil(t, dtoOrder)

	require.True(t, cartRepo.GetCalled)
	require.False(t, cartRepo.SaveCalled)
	require.False(t, orderRepo.CreateCalled)
	require.False(t, orderRepo.SaveCalled)
	require.False(t, productRepo.SaveCalled)
	require.False(t, productRepo.GetByIDsCalled)

}

func TestCreateFromCart_CartNotExists_ErrCartNotFound(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	tx := &MockTx{}
	cartRepo := &MockCartRepository{
		GetErr: errs.ErrCartNotFound,
	}
	orderRepo := &MockOrderRepository{}
	productRepo := &MockProductRepository{}

	service := order.NewService(order.ServiceDeps{
		CartRepository:    cartRepo,
		OrderRepository:   orderRepo,
		ProductRepository: productRepo,
		TxManager:         tx,
	})

	ctx := context.Background()
	dtoOrder, err := service.CreateFromCart(ctx, userID)
	require.ErrorIs(t, err, errs.ErrCartNotFound)

	require.Nil(t, dtoOrder)

	require.True(t, cartRepo.GetCalled)
	require.False(t, cartRepo.SaveCalled)
	require.False(t, orderRepo.CreateCalled)
	require.False(t, orderRepo.SaveCalled)
	require.False(t, productRepo.SaveCalled)
	require.False(t, productRepo.GetByIDsCalled)
}
