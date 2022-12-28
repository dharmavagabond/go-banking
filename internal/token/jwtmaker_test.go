package token

import (
	"strings"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

func TestJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker(randstr.String(32))
	require.NoError(t, err)

	username := strings.ToLower(randomdata.SillyName())
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestJWTExpiredToken(t *testing.T) {
	maker, err := NewJWTMaker(randstr.String(32))
	require.NoError(t, err)

	username := strings.ToLower(randomdata.SillyName())
	duration := -time.Minute
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.EqualError(t, err, ERR_EXPIRED_TOKEN.Error())
	require.Nil(t, payload)
}

func TestInvalidJWTSigningMethod(t *testing.T) {
	username := strings.ToLower(randomdata.SillyName())
	payload, err := NewPayload(username, time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := NewJWTMaker(randstr.String(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ERR_UNEXPECTED_JWT_SIGNING_METHOD.Error())
	require.Nil(t, payload)
}
