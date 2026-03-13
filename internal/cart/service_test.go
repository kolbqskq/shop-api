package cart_test

import (
	"context"
	"shop-api/internal/cart"
	"shop-api/internal/errs"
	"shop-api/internal/money"
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

type MockRepoCart struct {
	GetActiveCartCalled bool
	CreateCalled        bool
	SaveCalled          bool

	GetErr error

	CartSaved    *cart.Cart
	CartCreated  *cart.Cart
	CartToReturn *cart.Cart
}

func (m *MockRepoCart) GetActiveCart(ctx context.Context, userID uuid.UUID) (*cart.Cart, error) {
	m.GetActiveCartCalled = true
	return m.CartToReturn, m.GetErr
}
func (m *MockRepoCart) Create(ctx context.Context, cart *cart.Cart) error {
	m.CreateCalled = true
	m.CartCreated = cart
	return nil
}
func (m *MockRepoCart) Save(ctx context.Context, cart *cart.Cart) error {
	m.SaveCalled = true
	m.CartSaved = cart
	return nil
}

type MockRepoProduct struct {
	GetByIDsCalled bool

	GetErr           error
	ProductsToReturn []product.Product
}

func (m *MockRepoProduct) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]product.Product, error) {
	m.GetByIDsCalled = true
	return m.ProductsToReturn, m.GetErr
}

func (m *MockRepoProduct) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	return nil, nil
}

func TestGetActiveCart_ReturnsCartWithItems(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	productID, err := uuid.NewV7()
	require.NoError(t, err)
	c := cart.NewCart(id, userID)
	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	c.AddItem(productID, 1)

	repoCart := &MockRepoCart{
		CartToReturn: c,
	}
	repoProduct := &MockRepoProduct{
		ProductsToReturn: []product.Product{*p},
	}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.GetActiveCart(ctx, userID)
	require.NoError(t, err)

	require.True(t, repoCart.GetActiveCartCalled)
	require.True(t, repoProduct.GetByIDsCalled)

	require.Equal(t, c.ID(), dtoCart.ID)
	require.Len(t, dtoCart.Items, 1)
	require.Equal(t, productID, dtoCart.Items[0].ProductID)
	require.Equal(t, 1, dtoCart.Items[0].Quantity)
	require.Equal(t, int64(100), dtoCart.Items[0].Price.Amount)
	require.Equal(t, "test", dtoCart.Items[0].Name)
	require.True(t, dtoCart.Items[0].IsActive)
}

func TestGetActiveCart_NoCart_CreateNewCart(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)

	repoCart := &MockRepoCart{
		GetErr: errs.ErrCartNotFound,
	}
	repoProduct := &MockRepoProduct{}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.GetActiveCart(ctx, userID)
	require.NoError(t, err)

	require.True(t, repoCart.GetActiveCartCalled)
	require.True(t, repoCart.CreateCalled)

	require.False(t, repoProduct.GetByIDsCalled)

	require.NotNil(t, repoCart.CartCreated)
	require.NotNil(t, dtoCart)

	require.Len(t, dtoCart.Items, 0)
}

func TestAddToCart_ExistCart_AddItem(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)

	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	productID2, err := uuid.NewV7()
	require.NoError(t, err)

	c := cart.NewCart(id, userID)
	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	c.AddItem(productID, 1)

	p2, err := product.NewProduct(productID2, "test2", "test2", "test2", money.Money{Amount: 1000}, 100, true)
	require.NoError(t, err)

	repoCart := &MockRepoCart{
		CartToReturn: c,
	}
	repoProduct := &MockRepoProduct{
		ProductsToReturn: []product.Product{*p, *p2},
	}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.AddToCart(ctx, userID, productID2, 7)
	require.NoError(t, err)

	require.True(t, repoCart.GetActiveCartCalled)
	require.True(t, repoProduct.GetByIDsCalled)
	require.True(t, repoCart.SaveCalled)

	require.NotNil(t, repoCart.CartSaved)
	require.NotNil(t, dtoCart)

	require.Len(t, dtoCart.Items, 2)
	require.Len(t, repoCart.CartSaved.Items(), 2)

	require.ElementsMatch(t, []uuid.UUID{productID, productID2}, []uuid.UUID{dtoCart.Items[0].ProductID, dtoCart.Items[1].ProductID})
	require.ElementsMatch(t, []int{1, 7}, []int{dtoCart.Items[0].Quantity, dtoCart.Items[1].Quantity})
	require.ElementsMatch(t, []int64{100, 1000}, []int64{dtoCart.Items[0].Price.Amount, dtoCart.Items[1].Price.Amount})

	require.ElementsMatch(t, []uuid.UUID{productID, productID2}, []uuid.UUID{repoCart.CartSaved.Items()[0].ProductID, repoCart.CartSaved.Items()[1].ProductID})
	require.ElementsMatch(t, []int{1, 7}, []int{repoCart.CartSaved.Items()[0].Quantity, repoCart.CartSaved.Items()[1].Quantity})
}

func TestAddToCart_NoCart_AddItem(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)

	repoCart := &MockRepoCart{
		GetErr: errs.ErrCartNotFound,
	}
	repoProduct := &MockRepoProduct{
		ProductsToReturn: []product.Product{*p},
	}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.AddToCart(ctx, userID, productID, 7)
	require.NoError(t, err)

	require.True(t, repoCart.GetActiveCartCalled)
	require.True(t, repoProduct.GetByIDsCalled)
	require.True(t, repoCart.SaveCalled)
	require.True(t, repoCart.CreateCalled)

	require.NotNil(t, repoCart.CartSaved)
	require.NotNil(t, repoCart.CartCreated)
	require.NotNil(t, dtoCart)

	require.Len(t, dtoCart.Items, 1)
	require.Len(t, repoCart.CartSaved.Items(), 1)
}

func TestDecreaseFromCart_CartExist_DecreaseItem(t *testing.T) {
	cartID, err := uuid.NewV7()
	require.NoError(t, err)

	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	c := cart.NewCart(cartID, userID)
	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	c.AddItem(productID, 10)

	repoCart := &MockRepoCart{
		CartToReturn: c,
	}
	repoProduct := &MockRepoProduct{
		ProductsToReturn: []product.Product{*p},
	}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.UpdateFromCart(ctx, userID, productID, 7)
	require.NoError(t, err)

	require.True(t, repoCart.GetActiveCartCalled)
	require.True(t, repoProduct.GetByIDsCalled)
	require.True(t, repoCart.SaveCalled)
	require.False(t, repoCart.CreateCalled)

	require.NotNil(t, repoCart.CartSaved)
	require.NotNil(t, dtoCart)

	require.Len(t, dtoCart.Items, 1)
	require.Len(t, repoCart.CartSaved.Items(), 1)

	require.Equal(t, 3, dtoCart.Items[0].Quantity)
	require.Equal(t, cartID, dtoCart.ID)
	require.Equal(t, productID, dtoCart.Items[0].ProductID)

	require.Equal(t, 3, repoCart.CartSaved.Items()[0].Quantity)
	require.Equal(t, cartID, repoCart.CartSaved.ID())
	require.Equal(t, productID, repoCart.CartSaved.Items()[0].ProductID)
}

func TestDecreaseFromCart_CartExist_RemoveItem(t *testing.T) {
	cartID, err := uuid.NewV7()
	require.NoError(t, err)

	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	c := cart.NewCart(cartID, userID)
	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	c.AddItem(productID, 10)

	repoCart := &MockRepoCart{
		CartToReturn: c,
	}
	repoProduct := &MockRepoProduct{
		ProductsToReturn: []product.Product{*p},
	}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.UpdateFromCart(ctx, userID, productID, 10)
	require.NoError(t, err)

	require.True(t, repoCart.GetActiveCartCalled)
	require.False(t, repoProduct.GetByIDsCalled)
	require.True(t, repoCart.SaveCalled)
	require.False(t, repoCart.CreateCalled)

	require.NotNil(t, repoCart.CartSaved)
	require.NotNil(t, dtoCart)

	require.Len(t, dtoCart.Items, 0)
	require.Len(t, repoCart.CartSaved.Items(), 0)
}

func TestDecreaseFromCart_CartExist_InvalidQuantity(t *testing.T) {
	cartID, err := uuid.NewV7()
	require.NoError(t, err)

	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	c := cart.NewCart(cartID, userID)
	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	c.AddItem(productID, 10)

	repoCart := &MockRepoCart{
		CartToReturn: c,
	}
	repoProduct := &MockRepoProduct{
		ProductsToReturn: []product.Product{*p},
	}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.UpdateFromCart(ctx, userID, productID, -7)
	require.ErrorIs(t, err, errs.ErrInvalidQuantity)

	require.True(t, repoCart.GetActiveCartCalled)
	require.False(t, repoProduct.GetByIDsCalled)
	require.False(t, repoCart.SaveCalled)
	require.False(t, repoCart.CreateCalled)

	require.Nil(t, repoCart.CartSaved)
	require.Nil(t, dtoCart)
}

func TestDecreaseFromCart_NoCart_DecreaseItem(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	repoCart := &MockRepoCart{
		GetErr: errs.ErrCartNotFound,
	}
	repoProduct := &MockRepoProduct{}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.UpdateFromCart(ctx, userID, productID, 7)
	require.ErrorIs(t, err, errs.ErrCartNotFound)

	require.True(t, repoCart.GetActiveCartCalled)
	require.False(t, repoProduct.GetByIDsCalled)
	require.False(t, repoCart.SaveCalled)
	require.False(t, repoCart.CreateCalled)

	require.Nil(t, repoCart.CartSaved)
	require.Nil(t, dtoCart)
}

func TestRemoveFromCart_CartExist_RemoveItem(t *testing.T) {
	cartID, err := uuid.NewV7()
	require.NoError(t, err)

	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	c := cart.NewCart(cartID, userID)
	p, err := product.NewProduct(productID, "test", "test", "test", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	c.AddItem(productID, 10)

	repoCart := &MockRepoCart{
		CartToReturn: c,
	}
	repoProduct := &MockRepoProduct{
		ProductsToReturn: []product.Product{*p},
	}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.RemoveFromCart(ctx, userID, productID)
	require.NoError(t, err)

	require.True(t, repoCart.GetActiveCartCalled)
	require.False(t, repoProduct.GetByIDsCalled)
	require.True(t, repoCart.SaveCalled)

	require.NotNil(t, repoCart.CartSaved)
	require.NotNil(t, dtoCart)

	require.Len(t, dtoCart.Items, 0)
	require.Len(t, repoCart.CartSaved.Items(), 0)
}

func TestRemoveFromCart_CartExist_NotExistProduct(t *testing.T) {
	cartID, err := uuid.NewV7()
	require.NoError(t, err)

	userID, err := uuid.NewV7()
	require.NoError(t, err)

	productID, err := uuid.NewV7()
	require.NoError(t, err)

	c := cart.NewCart(cartID, userID)

	repoCart := &MockRepoCart{
		CartToReturn: c,
	}
	repoProduct := &MockRepoProduct{}

	tx := &MockTx{}

	service := cart.NewService(cart.ServiceDeps{
		Repository:        repoCart,
		ProductRepository: repoProduct,
		TxManager:         tx,
	})
	ctx := context.Background()
	dtoCart, err := service.RemoveFromCart(ctx, userID, productID)
	require.ErrorIs(t, err, errs.ErrProductNotFound)

	require.True(t, repoCart.GetActiveCartCalled)
	require.False(t, repoProduct.GetByIDsCalled)
	require.False(t, repoCart.SaveCalled)

	require.Nil(t, repoCart.CartSaved)
	require.Nil(t, dtoCart)

}
