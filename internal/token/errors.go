package token

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20"
)

var (
	ERR_UNEXPECTED_JWT_SIGNING_METHOD = errors.New("[Err]: Unexpected signing method")
	ERR_INVALID_JWT_KEY_SIZE          = fmt.Errorf(
		"[Err]: Invalid key size, must be at least %v characters long",
		MIN_SECRET_KEY_LENGTH,
	)
	ERR_INVALID_PASETO_KEY_SIZE = fmt.Errorf(
		"[Err]: Invalid key size, must be exactly %v characters long",
		chacha20.KeySize,
	)
	ERR_INVALID_PASETO_TOKEN    = errors.New("[Err]: Invalid token")
	ERR_EXPIRED_TOKEN           = errors.New("[Err]: Token has expired")
	ERR_CANT_CREATE_TOKEN_MAKER = errors.New("[Err]: Cannot create token maker")
)
