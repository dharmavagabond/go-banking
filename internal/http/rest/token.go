package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

type (
	renewAccessTokenRequest struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	renewAccessTokenResponse struct {
		AccessToken          string    `json:"access_token"`
		AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	}
)

func (server *Server) renewAccessToken(ectx echo.Context) (err error) {
	var (
		session             db.Session
		accessToken         string
		accessTokenPayload  *token.Payload
		refreshTokenPayload *token.Payload
		req                 = &renewAccessTokenRequest{}
	)

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	if refreshTokenPayload, err = server.tokenMaker.VerifyToken(req.RefreshToken); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if session, err = server.store.GetSession(ectx.Request().Context(), refreshTokenPayload.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if session.RefreshToken != req.RefreshToken {
		return echo.NewHTTPError(http.StatusUnauthorized, errors.New("Missmatched session token"))
	}

	if session.IsBlocked {
		return echo.NewHTTPError(http.StatusUnauthorized, errors.New("Blocked session"))
	}

	if session.Username != refreshTokenPayload.Username {
		return echo.NewHTTPError(http.StatusUnauthorized, errors.New("Incorrect session user"))
	}

	if time.Now().After(session.ExpiresAt) {
		return echo.NewHTTPError(http.StatusUnauthorized, errors.New("Expired session"))
	}

	if accessToken, accessTokenPayload, err = server.tokenMaker.CreateToken(
		refreshTokenPayload.Username,
		config.App.RefreshTokenDuration,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenPayload.ExpiredAt,
	}

	return ectx.JSON(http.StatusOK, res)
}
