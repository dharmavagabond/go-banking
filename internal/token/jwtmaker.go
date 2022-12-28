package token

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const MIN_SECRET_KEY_LENGTH = 32

type JWTMaker struct {
	secretKey string
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (token string, err error) {
	var payload *Payload

	if payload, err = NewPayload(username, duration); err != nil {
		return "", nil
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	if token, err = jwtToken.SignedString([]byte(maker.secretKey)); err != nil {
		return "", err
	}

	return token, nil
}

func (maker *JWTMaker) VerifyToken(token string) (payload *Payload, err error) {
	var (
		jwtToken *jwt.Token
		ok       bool
	)

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("%v: %v\n", ERR_UNEXPECTED_JWT_SIGNING_METHOD, token.Header["alg"])
			return nil, ERR_UNEXPECTED_JWT_SIGNING_METHOD
		}

		return []byte(maker.secretKey), nil
	}

	if jwtToken, err = jwt.ParseWithClaims(token, &Payload{}, keyFunc); err != nil {
		if verr, ok := err.(*jwt.ValidationError); ok && errors.Is(verr.Inner, ERR_EXPIRED_TOKEN) {
			err = ERR_EXPIRED_TOKEN
		} else {
			err = ERR_UNEXPECTED_JWT_SIGNING_METHOD
		}

		return nil, err
	}

	if payload, ok = jwtToken.Claims.(*Payload); !ok {
		return nil, ERR_UNEXPECTED_JWT_SIGNING_METHOD
	}

	return payload, nil
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < MIN_SECRET_KEY_LENGTH {
		return nil, ERR_INVALID_JWT_KEY_SIZE
	}

	return &JWTMaker{secretKey}, nil
}
