package product_test

import (
	"context"
	"shop-api/internal/core/errs"
	"shop-api/internal/core/money"
	"shop-api/internal/features/product"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type MockProductRepository struct {
	CreateCalled  bool
	GetByIDCalled bool
	SaveCalled    bool

	ProductToReturn *product.Product
	ProductCreated  *product.Product
	ProductSaved    *product.Product
}

func (m *MockProductRepository) Update(ctx context.Context, product *product.Product) error {
	m.SaveCalled = true
	m.ProductSaved = product
	return nil
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	m.GetByIDCalled = true
	return m.ProductToReturn, nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockProductRepository) Create(ctx context.Context, product *product.Product) error {
	m.CreateCalled = true
	m.ProductCreated = product
	return nil
}

func (m *MockProductRepository) List(ctx context.Context, filters product.ListFilters) ([]product.Product, error) {
	return nil, nil
}

func (m *MockProductRepository) Save(ctx context.Context, product *product.Product) error {
	m.SaveCalled = true
	m.ProductSaved = product
	return nil
}

func TestCreate_Success(t *testing.T) {
	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	req := product.CreateProductRequest{
		Name:        "test",
		Description: "test",
		Category:    "test",
		Price:       money.Money{Amount: 100},
		Stock:       10,
		IsActive:    true,
	}

	ctx := context.Background()

	res, err := service.CreateProduct(ctx, req)

	require.NoError(t, err)
	require.True(t, repo.CreateCalled)
	require.NotNil(t, repo.ProductCreated)
	require.NotNil(t, res)

	require.Equal(t, req.Name, repo.ProductCreated.Name())
	require.Equal(t, req.Description, repo.ProductCreated.Description())
	require.Equal(t, req.Category, repo.ProductCreated.Category())
	require.Equal(t, req.Price, repo.ProductCreated.Price())
	require.Equal(t, req.Stock, repo.ProductCreated.Stock())
	require.Equal(t, req.IsActive, repo.ProductCreated.IsActive())
}

func TestCreate_InvalidStock(t *testing.T) {
	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	req := product.CreateProductRequest{
		Name:        "test",
		Description: "test",
		Category:    "test",
		Price:       money.Money{Amount: 100},
		Stock:       -1,
		IsActive:    true,
	}
	ctx := context.Background()
	_, err := service.CreateProduct(ctx, req)

	require.ErrorIs(t, err, errs.ErrInvalidStock)
	require.False(t, repo.CreateCalled)
}

func TestCreate_InvalidPrice(t *testing.T) {
	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	req := product.CreateProductRequest{
		Name:        "test",
		Description: "test",
		Category:    "test",
		Price:       money.Money{Amount: -10},
		Stock:       10,
		IsActive:    true,
	}

	ctx := context.Background()
	_, err := service.CreateProduct(ctx, req)

	require.ErrorIs(t, err, errs.ErrInvalidPrice)
	require.False(t, repo.CreateCalled)
}

func TestChangeProduct_Success(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)
	passed, err := product.NewProduct(id, "passed", "passed", "passed", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	repo := &MockProductRepository{
		ProductToReturn: passed,
	}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	name := "changed"
	description := "changed"
	category := "changed"
	price := money.Money{Amount: 1}
	stock := 1
	isActive := false
	req := product.UpdateProductRequest{
		ID:          id,
		Name:        &name,
		Description: &description,
		Category:    &category,
		Price:       &price,
		Stock:       &stock,
		IsActive:    &isActive,
	}
	ctx := context.Background()
	dto, err := service.ChangeProduct(ctx, req)
	require.NoError(t, err)

	require.True(t, repo.GetByIDCalled)
	require.True(t, repo.SaveCalled)
	require.NotNil(t, repo.ProductSaved)

	require.Equal(t, name, repo.ProductSaved.Name())
	require.Equal(t, description, repo.ProductSaved.Description())
	require.Equal(t, category, repo.ProductSaved.Category())
	require.Equal(t, price.Amount, repo.ProductSaved.Price().Amount)
	require.Equal(t, stock, repo.ProductSaved.Stock())
	require.Equal(t, isActive, repo.ProductSaved.IsActive())

	require.Equal(t, id.String(), dto.ID)
	require.Equal(t, name, dto.Name)
	require.Equal(t, description, dto.Description)
	require.Equal(t, category, dto.Category)
	require.Equal(t, price.Amount, dto.Price)
	require.Equal(t, stock, dto.Available)

	require.Same(t, passed, repo.ProductSaved)
}

func TestChangeProduct_SuccessOnlyPrice(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)
	passed, err := product.NewProduct(id, "passed", "passed", "passed", money.Money{Amount: 100}, 10, true)
	require.NoError(t, err)
	repo := &MockProductRepository{
		ProductToReturn: passed,
	}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	price := money.Money{Amount: 1}
	req := product.UpdateProductRequest{
		ID:    id,
		Price: &price,
	}
	beforeName := passed.Name()
	beforeDescription := passed.Description()
	beforeCategory := passed.Category()
	beforeStock := passed.Stock()
	beforeIsActive := passed.IsActive()

	ctx := context.Background()
	dto, err := service.ChangeProduct(ctx, req)
	require.NoError(t, err)

	require.True(t, repo.GetByIDCalled)
	require.True(t, repo.SaveCalled)

	require.Equal(t, price.Amount, repo.ProductSaved.Price().Amount)

	require.Equal(t, beforeName, repo.ProductSaved.Name())
	require.Equal(t, beforeDescription, repo.ProductSaved.Description())
	require.Equal(t, beforeCategory, repo.ProductSaved.Category())
	require.Equal(t, beforeStock, repo.ProductSaved.Stock())
	require.Equal(t, beforeIsActive, repo.ProductSaved.IsActive())

	require.Equal(t, price.Amount, dto.Price)

	require.Equal(t, id.String(), dto.ID)
	require.Equal(t, beforeName, dto.Name)
	require.Equal(t, beforeDescription, dto.Description)
	require.Equal(t, beforeCategory, dto.Category)
	require.Equal(t, beforeStock, dto.Available)

	require.Same(t, passed, repo.ProductSaved)
}

func TestChangeProduct_NoFieldsToUpdate(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)

	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	req := product.UpdateProductRequest{
		ID: id,
	}

	ctx := context.Background()
	_, err = service.ChangeProduct(ctx, req)
	require.ErrorIs(t, err, errs.ErrNothingToUpdate)

	require.False(t, repo.GetByIDCalled)
	require.False(t, repo.SaveCalled)
}

func TestChangeProduct_NoID(t *testing.T) {
	repo := &MockProductRepository{}
	service := product.NewService(product.ServiceDeps{
		Repository: repo,
	})
	req := product.UpdateProductRequest{}

	ctx := context.Background()
	_, err := service.ChangeProduct(ctx, req)
	require.ErrorIs(t, err, errs.ErrMissingID)

	require.False(t, repo.GetByIDCalled)
	require.False(t, repo.SaveCalled)
}
