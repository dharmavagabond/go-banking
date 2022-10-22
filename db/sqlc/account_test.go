package db

import (
	"context"
	"testing"
	"time"

	"github.com/dharmavagabond/simple-bank/util"
	"github.com/jackc/pgx/v4"
	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(arg *CreateAccountParams) (Account, error) {
	if arg == nil {
		arg = &CreateAccountParams{
			Owner:    randomdata.SillyName(),
			Balance:  util.RandomMoney(),
			Currency: randomdata.Currency(),
		}
	}

	return testQueries.CreateAccount(context.Background(), *arg)
}

func TestCreateAccount(t *testing.T) {
	arg := &CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := createRandomAccount(arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
}

func TestGetAccount(t *testing.T) {
	account, err := createRandomAccount(nil)
	require.NoError(t, err)
	account2, err := testQueries.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account.ID, account2.ID)
	require.Equal(t, account.Owner, account2.Owner)
	require.Equal(t, account.Balance, account2.Balance)
	require.Equal(t, account.Currency, account2.Currency)
	require.WithinDuration(t, account.CreatedAt, account2.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	account, err := createRandomAccount(nil)
	require.NoError(t, err)
	arg := UpdateAccountParams{
		ID:      account.ID,
		Balance: util.RandomMoney(),
	}

	account2, err := testQueries.UpdateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, account2)
	require.Equal(t, account.Owner, account2.Owner)
	require.Equal(t, arg.Balance, account2.Balance)
	require.Equal(t, account.Currency, account2.Currency)
	require.WithinDuration(t, account.CreatedAt, account2.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	account, err := createRandomAccount(nil)
	require.NoError(t, err)
	err = testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)
	account2, err := testQueries.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, account2)
}

func TestListAccount(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(nil)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
