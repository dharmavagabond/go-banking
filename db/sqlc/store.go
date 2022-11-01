package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Store struct {
	*Queries
	db *pgxpool.Pool
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	var (
		err error
		tx  pgx.Tx
	)

	if tx, err = store.db.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		return err
	}

	if txErr := fn(New(tx)); txErr != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("[Tx Err]: %v\n[Rollback Err]: %v\n", txErr, rbErr)
		} else {
			return txErr
		}
	}

	return tx.Commit(ctx)
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	txError := store.execTx(ctx, func(q *Queries) error {
		var err error
		if result.Transfer, err = q.CreateTransfer(
			ctx,
			CreateTransferParams{
				FromAccountID: arg.FromAccountID,
				ToAccountID:   arg.ToAccountID,
				Amount:        arg.Amount,
			},
		); err != nil {
			return err
		}

		if result.FromEntry, err = q.CreateEntry(
			ctx,
			CreateEntryParams{
				AccountID: sql.NullInt64{Int64: arg.FromAccountID, Valid: true},
				Amount:    -arg.Amount,
			},
		); err != nil {
			return err
		}

		if result.ToEntry, err = q.CreateEntry(
			ctx,
			CreateEntryParams{
				AccountID: sql.NullInt64{Int64: arg.ToAccountID, Valid: true},
				Amount:    arg.Amount,
			},
		); err != nil {
			return err
		}

		if result.FromAccount, err = q.AddAccountBalance(
			ctx,
			AddAccountBalanceParams{
				ID:               arg.FromAccountID,
				AmountToTransfer: -arg.Amount,
			},
		); err != nil {
			return err
		}

		result.ToAccount, err = q.AddAccountBalance(
			ctx,
			AddAccountBalanceParams{
				ID:               arg.ToAccountID,
				AmountToTransfer: arg.Amount,
			},
		)

		return err
	})

	return result, txError
}
