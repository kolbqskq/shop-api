package product_test

import (
	"context"
	"shop-api/internal/money"
	"shop-api/internal/product"
	"shop-api/pkg/database"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestConcurrentReserve(t *testing.T) {

	dbPool, err := database.CreateTestDbPool()
	if err != nil {
		t.Fatal("failed to init db")
	}
	productRepo := product.NewRepository(product.RepositoryDeps{
		DbPool: dbPool,
	})
	ctx := context.Background()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	productRepo.Create(ctx, &product.Product{
		ID:          id,
		Name:        "test",
		Description: "test",
		Category:    "test",
		Price:       money.Money{Amount: 100},
		Stock:       10,
		IsActive:    true,
	})
	var wg sync.WaitGroup

	count := 100

	wg.Add(count)

	start := make(chan struct{})

	errs := make([]error, count)

	for i := range count {
		go func() {
			defer wg.Done()
			<-start
			errs[i] = productRepo.Reserve(ctx, []product.Reservation{
				{
					ProductID: id,
					Quantity:  7,
				},
			})
		}()
	}

	close(start)
	wg.Wait()

	success := 0

	for _, err := range errs {
		if err == nil {
			success++
		}
	}

	if success != 1 {
		t.Fatalf("expected one success, got %d", success)
	}
}
