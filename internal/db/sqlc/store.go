package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Store interface {
	Querier
	TransferTx(context.Context, CreateTransferParams) (TransferTxResult, error)
	CreateUserTx(
		context.Context,
		CreateUserTxParams,
	) (CreateUserTxResult, error)
}

type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

var (
	once  sync.Once
	store Store
)

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func NewStore() Store {
	once.Do(func() {
		var (
			dbconfig *pgxpool.Config
			dbpool   *pgxpool.Pool
			err      error
		)

		logger := echo.New().Logger

		if dbconfig, err = pgxpool.ParseConfig(config.Postgres.DSN); err != nil {
			logger.Fatal("[Err]: ", err)
		}

		if dbpool, err = pgxpool.NewWithConfig(context.Background(), dbconfig); err != nil {
			logger.Fatal("[Err]: ", err)
		}

		if err = dbpool.Ping(context.Background()); err != nil {
			logger.Fatal("[Err]: ", err)
		}

		store = &SQLStore{
			db:      dbpool,
			Queries: New(dbpool),
		}
	})

	return store
}

func (store *SQLStore) execTx(
	ctx context.Context,
	fn func(*Queries) error,
) error {
	var (
		err error
		tx  pgx.Tx
	)

	if tx, err = store.db.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		return err
	}

	if txErr := fn(New(tx)); txErr != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf(
				"[Tx Err]: %v\n[Rollback Err]: %v\n",
				txErr,
				rbErr,
			)
		} else {
			return txErr
		}
	}

	return tx.Commit(ctx)
}

func (store *SQLStore) TransferTx(
	ctx context.Context,
	arg CreateTransferParams,
) (result TransferTxResult, txError error) {
	txError = store.execTx(ctx, func(q *Queries) (err error) {
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
				AccountID: pgtype.Int8{Int64: arg.FromAccountID, Valid: true},
				Amount:    -arg.Amount,
			},
		); err != nil {
			return err
		}

		if result.ToEntry, err = q.CreateEntry(
			ctx,
			CreateEntryParams{
				AccountID: pgtype.Int8{Int64: arg.ToAccountID, Valid: true},
				Amount:    arg.Amount,
			},
		); err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = transferMoney(
				ctx,
				q,
				arg.FromAccountID,
				arg.ToAccountID,
				-arg.Amount,
			)
		} else {
			result.ToAccount, result.FromAccount, err = transferMoney(ctx, q, arg.ToAccountID, arg.FromAccountID, arg.Amount)
		}

		return err
	})

	return result, txError
}

func transferMoney(
	ctx context.Context,
	q *Queries,
	fromAccountID,
	toAccountID,
	amountToTransfer int64,
) (fromAccount, toAccount Account, err error) {
	fromAccount, err = q.AddAccountBalance(
		ctx,
		AddAccountBalanceParams{
			ID:               fromAccountID,
			AmountToTransfer: amountToTransfer,
		},
	)

	if err != nil {
		return
	}

	toAccount, err = q.AddAccountBalance(
		ctx,
		AddAccountBalanceParams{
			ID:               toAccountID,
			AmountToTransfer: amountToTransfer * -1,
		},
	)

	return
}
