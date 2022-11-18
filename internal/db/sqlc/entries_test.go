package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dharmavagabond/simple-bank/internal/util"
	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {
	account, _ := createRandomAccount(nil)
	arg := CreateEntryParams{
		AccountID: sql.NullInt64{Int64: account.ID, Valid: true},
		Amount:    util.RandomMoney(),
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	id, err := arg.AccountID.Value()
	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.Equal(t, id, account.ID)
	require.Equal(t, arg.Amount, entry.Amount)
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)
}

func TestGetEntry(t *testing.T) {
	account, _ := createRandomAccount(nil)
	arg := CreateEntryParams{
		AccountID: sql.NullInt64{Int64: account.ID, Valid: true},
		Amount:    util.RandomMoney(),
	}
	entry, _ := testQueries.CreateEntry(context.Background(), arg)
	entry2, err := testQueries.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)
	require.Equal(t, entry.ID, entry2.ID)
	require.Equal(t, entry.AccountID, entry2.AccountID)
	require.Equal(t, entry.Amount, entry2.Amount)
	require.WithinDuration(t, entry.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	account, _ := createRandomAccount(nil)

	for i := 0; i < 10; i++ {
		_, _ = testQueries.CreateEntry(
			context.Background(),
			CreateEntryParams{
				AccountID: sql.NullInt64{Int64: account.ID, Valid: true},
				Amount:    util.RandomMoney(),
			},
		)
	}

	arg := ListEntriesParams{
		AccountID: sql.NullInt64{Int64: account.ID, Valid: true},
		Limit:     5,
		Offset:    5,
	}
	entries, err := testQueries.ListEntries(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.Equal(t, arg.AccountID, entry.AccountID)
	}
}
