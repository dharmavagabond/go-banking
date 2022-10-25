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
	amountToTransfer := int64(10)

	for i := 0; i < 5; i++ {
		go func() {
			result, err := store.TransferTx(
				ctx,
				TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        amountToTransfer,
				},
			)

			errorsch <- err
			resultsch <- result
		}()
	}

	for i := 0; i < 5; i++ {
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

		// TODO: check if balance is updated
	}
}
