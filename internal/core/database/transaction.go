package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbTransaction struct {
	Tx pgx.Tx
}

func (d *DbTransaction) Commit(ctx context.Context) error {
	if err := d.Tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

func (d *DbTransaction) Rollback(ctx context.Context) error {
	if err := d.Tx.Rollback(ctx); err != nil {
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

func (d *DbTransactionManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := d.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err //Добавить лог
	}

	txCtx := InjectTx(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx) //Добавить лог
			panic(r)
		}
	}()
	if err := fn(txCtx); err != nil {
		tx.Rollback(ctx) //Добавить лог
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

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func Executor(ctx context.Context, pool DBTX) DBTX {
	if tx, ok := ExtractTx(ctx); ok {
		return tx
	}
	return pool
}
