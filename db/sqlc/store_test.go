package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(dbpool)
	ctx := context.Background()
	fromAccount, _ := createRandomAccount(nil)
	toAccount, _ := createRandomAccount(nil)
	errorsch := make(chan error)
	resultsch := make(chan TransferTxResult)
	executedTransactions := 5
	amountToTransfer := int64(10)
	existed := make(map[int]bool)

	for i := 0; i < executedTransactions; i++ {
		go func() {
			result, err := store.TransferTx(
				ctx,
				CreateTransferParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        amountToTransfer,
				},
			)

			errorsch <- err
			resultsch <- result
		}()
	}

	for i := 0; i < executedTransactions; i++ {
		var err error

		err = <-errorsch
		require.NoError(t, err)

		result := <-resultsch
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)
		require.Equal(t, fromAccount.ID, transfer.FromAccountID)
		require.Equal(t, toAccount.ID, transfer.ToAccountID)
		require.Equal(t, amountToTransfer, transfer.Amount)

		_, err = store.GetTransfer(ctx, transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		fromEntryAccountId, _ := fromEntry.AccountID.Value()
		require.NotEmpty(t, fromEntry)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		require.Equal(t, fromAccount.ID, fromEntryAccountId)
		require.Equal(t, -amountToTransfer, fromEntry.Amount)

		_, err = store.GetEntry(ctx, fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		toEntryAccountId, _ := toEntry.AccountID.Value()
		require.NotEmpty(t, toEntry)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		require.Equal(t, toAccount.ID, toEntryAccountId)
		require.Equal(t, amountToTransfer, toEntry.Amount)

		_, err = store.GetEntry(ctx, toEntry.ID)
		require.NoError(t, err)

		uFromAccount := result.FromAccount
		require.NotEmpty(t, uFromAccount)
		require.Equal(t, fromAccount.ID, uFromAccount.ID)

		uToAccount := result.ToAccount
		require.NotEmpty(t, uToAccount)
		require.Equal(t, toAccount.ID, uToAccount.ID)

		diff1 := fromAccount.Balance - uFromAccount.Balance
		diff2 := uToAccount.Balance - toAccount.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amountToTransfer == 0)

		k := int(diff1 / amountToTransfer)
		require.True(t, k >= 1 && k <= executedTransactions)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updatedFromAccount, err := testQueries.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)

	updatedToAccount, err := testQueries.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)

	require.Equal(t, fromAccount.Balance-int64(executedTransactions)*amountToTransfer, updatedFromAccount.Balance)
	require.Equal(t, toAccount.Balance+int64(executedTransactions)*amountToTransfer, updatedToAccount.Balance)
}

func TestTransferDeadlockTx(t *testing.T) {
	store := NewStore(dbpool)
	ctx := context.Background()
	fromAccount, _ := createRandomAccount(nil)
	toAccount, _ := createRandomAccount(nil)
	errorsch := make(chan error)
	executedTransactions := 10
	amountToTransfer := int64(10)

	for i := 0; i < executedTransactions; i++ {
		fromAccountID := fromAccount.ID
		toAccountID := toAccount.ID

		if i%2 == 1 {
			fromAccountID = toAccount.ID
			toAccountID = fromAccount.ID
		}

		go func() {
			_, err := store.TransferTx(
				ctx,
				CreateTransferParams{
					FromAccountID: fromAccountID,
					ToAccountID:   toAccountID,
					Amount:        amountToTransfer,
				},
			)

			errorsch <- err
		}()
	}

	for i := 0; i < executedTransactions; i++ {
		err := <-errorsch
		require.NoError(t, err)
	}

	updatedFromAccount, err := testQueries.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)

	updatedToAccount, err := testQueries.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)

	require.Equal(t, fromAccount.Balance, updatedFromAccount.Balance)
	require.Equal(t, toAccount.Balance, updatedToAccount.Balance)
}
