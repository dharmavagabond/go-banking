package db

import (
	"context"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/require"
)

func createRandomUser(arg *CreateUserParams) (User, error) {
	if arg == nil {

		arg = &CreateUserParams{
			Username:       randomdata.SillyName(),
			HashedPassword: "secret",
			FullName:       randomdata.FullName(randomdata.RandomGender),
			Email:          randomdata.Email(),
		}
	}

	return testQueries.CreateUser(context.Background(), *arg)
}

func TestCreateUser(t *testing.T) {
	arg := &CreateUserParams{
		Username:       randomdata.SillyName(),
		HashedPassword: "secret",
		FullName:       randomdata.FullName(randomdata.RandomGender),
		Email:          randomdata.Email(),
	}

	user, err := createRandomUser(arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)
}

func TestGetUser(t *testing.T) {
	createdUser, _ := createRandomUser(nil)
	fetchedUser, err := testQueries.GetUser(context.Background(), createdUser.Username)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedUser)
	require.Equal(t, createdUser.Username, fetchedUser.Username)
	require.Equal(t, createdUser.HashedPassword, fetchedUser.HashedPassword)
	require.Equal(t, createdUser.Email, fetchedUser.Email)
	require.WithinDuration(t, createdUser.PasswordChangedAt, fetchedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, createdUser.CreatedAt, fetchedUser.CreatedAt, time.Second)
}
