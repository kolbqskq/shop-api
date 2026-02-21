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

func newTestCustomProduct(t *testing.T, name, desc, category string, stock int) *product.Product {
	t.Helper()

	id, err := uuid.NewV7()
	require.NoError(t, err)

	p, err := product.NewProduct(id, name, desc, category, money.Money{Amount: 100}, stock, true)
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

	prod := newTestCustomProduct(t, "test", "test", "test", stock)

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

	prod1 := newTestCustomProduct(t, "test", "test", category, 10)
	require.NoError(t, repo.Create(ctx, prod1))

	prod2 := newTestCustomProduct(t, "test", "test", otherCategory, 10)
	require.NoError(t, repo.Create(ctx, prod2))

	products, err := repo.List(ctx, product.ListFilters{
		Limit:    10,
		Offset:   0,
		SortBy:   product.SortByCreatedAt,
		SortDesc: false,
		Category: &category,
	})
	require.NoError(t, err)

	require.Len(t, products, 1)
	require.Equal(t, category, products[0].Category)
}
