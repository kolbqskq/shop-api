package product_test

import (
	"context"
	"shop-api/internal/errs"
	"shop-api/internal/money"
	"shop-api/internal/product"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type MockProductRepository struct {
	CreateCalled  bool
	ProductPassed *product.Product
}

func (m *MockProductRepository) Update(ctx context.Context, product *product.Product) error {
	return nil
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	return nil, nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockProductRepository) Create(ctx context.Context, product *product.Product) error {
	m.CreateCalled = true
	m.ProductPassed = product
	return nil
}

func (m *MockProductRepository) List(ctx context.Context, filters product.ListFilters) ([]product.Product, error) {
	return nil, nil
}

func TestCreate_Success(t *testing.T) {
	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	name := "test"
	price := money.Money{Amount: 100}
	stock := 10
	ctx := context.Background()
	product, err := service.CreateProduct(ctx, product.CreateProductRequest{
		Name:        name,
		Description: "test",
		Category:    "test",
		Price:       price,
		Stock:       stock,
		IsActive:    true,
	})

	require.NoError(t, err)
	require.Equal(t, product, repo.ProductPassed)
	require.True(t, repo.CreateCalled)
	require.NotNil(t, repo.ProductPassed)
	require.Equal(t, name, product.Name)
	require.Equal(t, price.Amount, product.Price.Amount)
	require.Equal(t, stock, product.Stock)
}

func TestCreate_InvalidStock(t *testing.T) {
	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	name := "test"
	price := money.Money{Amount: 100}
	stock := -1
	ctx := context.Background()
	_, err := service.CreateProduct(ctx, product.CreateProductRequest{
		Name:        name,
		Description: "test",
		Category:    "test",
		Price:       price,
		Stock:       stock,
		IsActive:    true,
	})

	require.ErrorIs(t, err, errs.ErrInvalidStock)
	require.False(t, repo.CreateCalled)
}

func TestCreate_InvalidPrice(t *testing.T) {
	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	name := "test"
	price := money.Money{Amount: -10}
	stock := 10
	ctx := context.Background()
	_, err := service.CreateProduct(ctx, product.CreateProductRequest{
		Name:        name,
		Description: "test",
		Category:    "test",
		Price:       price,
		Stock:       stock,
		IsActive:    true,
	})

	require.ErrorIs(t, err, errs.ErrInvalidPrice)
	require.False(t, repo.CreateCalled)
}
