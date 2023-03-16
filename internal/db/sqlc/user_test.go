package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/alexedwards/argon2id"
	"github.com/stretchr/testify/require"
)

func createRandomUser(arg *CreateUserParams) (User, error) {
	if arg == nil {
		hash, err := argon2id.CreateHash(randomdata.Alphanumeric(16), argon2id.DefaultParams)

		if err != nil {
			return User{}, err
		}

		arg = &CreateUserParams{
			Username:       strings.ToLower(randomdata.SillyName()),
			HashedPassword: hash,
			FullName:       randomdata.FullName(randomdata.RandomGender),
			Email:          randomdata.Email(),
		}
	}

	return testQueries.CreateUser(context.Background(), *arg)
}

func TestCreateUser(t *testing.T) {
	hash, err := argon2id.CreateHash(randomdata.SillyName(), argon2id.DefaultParams)
	require.NoError(t, err)

	arg := &CreateUserParams{
		Username:       strings.ToLower(randomdata.SillyName()),
		HashedPassword: hash,
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

func TestUpdate(t *testing.T) {
	oldUser, _ := createRandomUser(nil)
	newFullName := randomdata.FullName(randomdata.RandomGender)
	updatedParams := UpdateUserParams{
		Username: oldUser.Username,
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), updatedParams)
	require.NoError(t, err)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}
