package db

import (
	"context"
	"testing"
	"time"

	"github.com/dharmavagabond/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(fromAccountId, toAccountId int64, arg *CreateTransferParams) (Transfer, error) {
	if arg == nil {
		arg = &CreateTransferParams{
			FromAccountID: fromAccountId,
			ToAccountID:   toAccountId,
			Amount:        util.RandomMoney(),
		}
	}

	return testQueries.CreateTransfer(context.Background(), *arg)
}

func TestCreateTransfer(t *testing.T) {
	fromAccount, _ := createRandomAccount(nil)
	toAccount, _ := createRandomAccount(nil)
	arg := &CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
	}
	transfer, err := createRandomTransfer(fromAccount.ID, toAccount.ID, arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
}

func TestGetTransfer(t *testing.T) {
	fromAccount, _ := createRandomAccount(nil)
	toAccount, _ := createRandomAccount(nil)
	transfer, _ := createRandomTransfer(fromAccount.ID, toAccount.ID, nil)
	fetchedTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedTransfer)
	require.Equal(t, transfer.ID, fetchedTransfer.ID)
	require.Equal(t, transfer.FromAccountID, fetchedTransfer.FromAccountID)
	require.Equal(t, transfer.ToAccountID, fetchedTransfer.ToAccountID)
	require.Equal(t, transfer.Amount, fetchedTransfer.Amount)
	require.WithinDuration(t, transfer.CreatedAt, fetchedTransfer.CreatedAt, time.Second)
}

func TestListTransfer(t *testing.T) {
	fromAccount, _ := createRandomAccount(nil)
	toAccount, _ := createRandomAccount(nil)

	for i := 0; i < 10; i++ {
		createRandomTransfer(fromAccount.ID, toAccount.ID, nil)
	}

	arg := ListTransfersParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Limit:         5,
		Offset:        5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
		require.True(t, transfer.FromAccountID == fromAccount.ID && transfer.ToAccountID == toAccount.ID)
	}
}
