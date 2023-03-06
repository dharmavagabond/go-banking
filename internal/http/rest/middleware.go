package rest

import (
	"net/http"

	"github.com/MadAppGang/httplog"
	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	AUTH_HEADER               = "Authorization"
	AUTH_TYPE_BEARER          = "Bearer"
	AUTHORIZATION_PAYLOAD_KEY = "authorizationPayloadKey"
)

var authMiddleware = middleware.KeyAuth(func(key string, ectx echo.Context) (bool, error) {
	var (
		tokenMaker token.Maker
		payload    *token.Payload
		err        error
	)

	if tokenMaker, err = token.NewPasetoMaker(config.App.TokenSymmetricKey); err != nil {
		return false, err
	}

	if payload, err = tokenMaker.VerifyToken(key); err != nil {
		return false, err
	}

	ectx.Set(AUTHORIZATION_PAYLOAD_KEY, payload)

	return true, nil
})

func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ectx echo.Context) error {
		logger := httplog.Logger(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if err := next(ectx); err != nil {
				ectx.Error(err)
			}

			lrw, _ := rw.(httplog.ResponseWriter)
			lrw.Set(ectx.Response().Status, int(ectx.Response().Size))
		}))
		logger.ServeHTTP(ectx.Response().Writer, ectx.Request())
		return nil
	}
}
