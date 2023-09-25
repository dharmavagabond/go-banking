package db

import (
	"context"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/dharmavagabond/simple-bank/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(arg *CreateAccountParams) (Account, error) {
	if arg == nil {
		user, _ := createRandomUser(nil)
		arg = &CreateAccountParams{
			Owner:    user.Username,
			Balance:  util.RandomMoney(),
			Currency: randomdata.Currency(),
		}
	}

	return testQueries.CreateAccount(context.Background(), *arg)
}

func TestCreateAccount(t *testing.T) {
	user, _ := createRandomUser(nil)
	arg := &CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: randomdata.Currency(),
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
	require.WithinDuration(t, account.CreatedAt.Time, account2.CreatedAt.Time, time.Second)
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
	require.WithinDuration(t, account.CreatedAt.Time, account2.CreatedAt.Time, time.Second)
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
	var lastAccount Account

	for i := 0; i < 10; i++ {
		lastAccount, _ = createRandomAccount(nil)
	}

	arg := ListAccountsParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
	}
}
