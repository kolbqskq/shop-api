package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbTransaction struct {
	Tx pgx.Tx
}

func (tx *DbTransaction) Commit(ctx context.Context) error {
	if err := tx.Tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

func (tx *DbTransaction) Rollback(ctx context.Context) error {
	if err := tx.Tx.Rollback(ctx); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	return nil
}

type DbTransactionManager struct {
	Db *pgxpool.Pool
}

func NewDbTransactionManager(db *pgxpool.Pool) DbTransactionManager {
	return DbTransactionManager{
		Db: db,
	}
}

func (tm *DbTransactionManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		}
	}()
	if err := fn(txCtx); err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

type txKey struct {
}

func InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func ExtractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
