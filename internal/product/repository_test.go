package product_test

import (
	"context"
	"shop-api/internal/database"
	"shop-api/internal/errs"
	"shop-api/internal/money"
	"shop-api/internal/product"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestCustomProduct(t *testing.T, name, desc, category string, stock int, price int64) *product.Product {
	t.Helper()

	id, err := uuid.NewV7()
	require.NoError(t, err)

	p, err := product.NewProduct(id, name, desc, category, money.Money{Amount: price}, stock, true)
	require.NoError(t, err)

	return p
}

func setupRepo(t *testing.T) (*product.Repository, context.Context) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	dbPool, err := database.CreateTestDbPool()
	require.NoError(t, err)

	repo := product.NewRepository(product.RepositoryDeps{
		DbPool: dbPool,
	})

	_, err = dbPool.Exec(ctx, "DELETE FROM products")
	require.NoError(t, err)

	return repo, ctx
}

func TestConcurrentReserve(t *testing.T) {

	count := 100
	stock := 21
	reserve := 7
	suc := stock / reserve

	repo, ctx := setupRepo(t)

	prod := newTestCustomProduct(t, "test", "test", "test", stock, 100)

	err := repo.Create(ctx, prod)
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Add(count)

	start := make(chan struct{})

	errors := make([]error, count)

	for i := range count {
		i := i
		go func() {
			defer wg.Done()
			<-start

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, errors[i] = repo.Reserve(ctx, []product.Reservation{
				{
					ProductID: prod.ID,
					Quantity:  reserve,
				},
			})
		}()
	}

	close(start)
	wg.Wait()

	success := 0

	for _, err := range errors {
		if err == nil {
			success++
		} else {
			require.ErrorIs(t, err, errs.ErrNotEnoughStock)
		}

	}
	require.Equal(t, suc, success)

	product, err := repo.GetByID(ctx, prod.ID)
	require.NoError(t, err)

	require.Equal(t, reserve*suc, product.Reserved)
}

func TestList_FilterByCategory(t *testing.T) {
	repo, ctx := setupRepo(t)

	category := "games"
	otherCategory := "books"

	prod1 := newTestCustomProduct(t, "test", "test", category, 10, 100)
	require.NoError(t, repo.Create(ctx, prod1))

	prod2 := newTestCustomProduct(t, "test", "test", otherCategory, 10, 100)
	require.NoError(t, repo.Create(ctx, prod2))

	products, err := repo.List(ctx, product.ListFilters{
		Limit:    10,
		Offset:   0,
		SortBy:   product.SortByStock,
		SortDesc: false,
		Category: &category,
	})
	require.NoError(t, err)

	require.Len(t, products, 1)
	require.Equal(t, category, products[0].Category)
}

func TestList_FilterLimit(t *testing.T) {
	repo, ctx := setupRepo(t)

	count := 10
	limit := 3

	for range count {
		require.NoError(t, repo.Create(ctx, newTestCustomProduct(t, "test", "test", "test", 10, 100)))
	}

	products, err := repo.List(ctx, product.ListFilters{
		Limit:    limit,
		Offset:   0,
		SortBy:   product.SortByCreatedAt,
		SortDesc: false,
	})
	require.NoError(t, err)

	require.Len(t, products, 3)
}

func TestList_FilterOffset(t *testing.T) {
	repo, ctx := setupRepo(t)

	count := 10
	offset := 3

	for i := range count {
		require.NoError(t, repo.Create(ctx, newTestCustomProduct(t, "test", "test", "test", i, 100)))
	}

	products, err := repo.List(ctx, product.ListFilters{
		Limit:    1,
		Offset:   offset,
		SortBy:   product.SortByStock,
		SortDesc: false,
	})
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.Equal(t, offset, products[0].Stock)
}

func TestList_FilterDescTrue(t *testing.T) {
	repo, ctx := setupRepo(t)

	count := 10
	sortDesc := true

	for i := range count {
		require.NoError(t, repo.Create(ctx, newTestCustomProduct(t, "test", "test", "test", i, 100)))
	}

	products, err := repo.List(ctx, product.ListFilters{
		Limit:    10,
		Offset:   0,
		SortBy:   product.SortByStock,
		SortDesc: sortDesc,
	})
	require.NoError(t, err)
	require.Len(t, products, 10)
	require.Equal(t, count-1, products[0].Stock)
}

func TestList_FilterMinPrice(t *testing.T) {
	repo, ctx := setupRepo(t)

	count := 10
	minPrice := int64(500)

	for i := 0; i < count*100; i += 100 {
		require.NoError(t, repo.Create(ctx, newTestCustomProduct(t, "test", "test", "test", 10, int64(i))))
	}

	products, err := repo.List(ctx, product.ListFilters{
		Limit:    10,
		Offset:   0,
		SortBy:   product.SortByStock,
		SortDesc: false,

		MinPrice: &minPrice,
	})
	require.NoError(t, err)

	for i := range products {
		require.GreaterOrEqual(t, products[i].Price.Amount, minPrice)
	}
}

func TestList_FilterMaxPrice(t *testing.T) {
	repo, ctx := setupRepo(t)

	count := 10
	maxPrice := int64(500)

	for i := 0; i < count*100; i += 100 {
		require.NoError(t, repo.Create(ctx, newTestCustomProduct(t, "test", "test", "test", 10, int64(i))))
	}

	products, err := repo.List(ctx, product.ListFilters{
		Limit:    10,
		Offset:   0,
		SortBy:   product.SortByStock,
		SortDesc: false,

		MaxPrice: &maxPrice,
	})
	require.NoError(t, err)

	for i := range products {
		require.LessOrEqual(t, products[i].Price.Amount, maxPrice)
	}
}
