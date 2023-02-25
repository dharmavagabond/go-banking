package token

import (
	"strings"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(randstr.String(32))
	require.NoError(t, err)

	username := strings.ToLower(randomdata.SillyName())
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)
	token, payload, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestPasetoExpiredToken(t *testing.T) {
	maker, err := NewPasetoMaker(randstr.String(32))
	require.NoError(t, err)

	username := strings.ToLower(randomdata.SillyName())
	duration := -time.Minute
	token, _, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.EqualError(t, err, ERR_EXPIRED_TOKEN.Error())
	require.Nil(t, payload)
	require.Nil(t, payload)
}
