package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type (
	createUserRequest struct {
		Username string `json:"username"  validate:"required,alphanum"`
		Password string `json:"password"  validate:"required,min=10"`
		FullName string `json:"full_name" validate:"required"`
		Email    string `json:"email"     validate:"required,email"`
	}
	userResponse struct {
		Username          string    `json:"username"`
		FullName          string    `json:"full_name"`
		Email             string    `json:"email"`
		PasswordChangedAt time.Time `json:"password_changed_at"`
		CreatedAt         time.Time `json:"created_at"`
	}
	loginUserRequest struct {
		Username string `json:"username" validate:"required,alphanum"`
		Password string `json:"password" validate:"required,min=10"`
	}
	loginUserResponse struct {
		SessionID             uuid.UUID    `json:"session_id"`
		AccessToken           string       `json:"access_token"`
		AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
		RefreshToken          string       `json:"refresh_token"`
		RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
		User                  userResponse `json:"user"`
	}
)

var argonParams = &argon2id.Params{
	Memory:      128 * 1024,
	Iterations:  4,
	Parallelism: 4,
	SaltLength:  128,
	KeyLength:   128,
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt.Time,
		CreatedAt:         user.CreatedAt.Time,
	}
}

func (server *Server) createUser(ectx echo.Context) (err error) {
	var (
		user         db.User
		hashPassword string
		req          = &createUserRequest{}
	)

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	if hashPassword, err = argon2id.CreateHash(req.Password, argonParams); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	if user, err = server.store.CreateUser(ectx.Request().Context(), arg); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			}
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, newUserResponse(user))
}

func (server *Server) loginUser(ectx echo.Context) (err error) {
	var (
		session             db.Session
		accessToken         string
		accessTokenPayload  *token.Payload
		refreshToken        string
		refreshTokenPayload *token.Payload
		user                db.User
		ok                  bool
		req                 = &loginUserRequest{}
	)

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	if user, err = server.store.GetUser(ectx.Request().Context(), req.Username); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if ok, err = argon2id.ComparePasswordAndHash(req.Password, user.HashedPassword); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Contraseña incorrecta.")
	}

	if accessToken, accessTokenPayload, err = server.tokenMaker.CreateToken(user.Username, config.App.AccessTokenDuration); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if refreshToken, refreshTokenPayload, err = server.tokenMaker.CreateToken(
		req.Username,
		config.App.RefreshTokenDuration,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if session, err = server.store.CreateSession(ectx.Request().Context(), db.CreateSessionParams{
		ID:           pgtype.UUID{Bytes: refreshTokenPayload.ID, Valid: true},
		Username:     req.Username,
		RefreshToken: refreshToken,
		UserAgent:    ectx.Request().UserAgent(),
		ClientIp:     ectx.RealIP(),
		ExpiresAt:    pgtype.Timestamptz{Time: refreshTokenPayload.ExpiredAt, Valid: true},
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sessionID, err := uuid.FromBytes(session.ID.Bytes[:])
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := loginUserResponse{
		SessionID:             sessionID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	return ectx.JSON(http.StatusOK, res)
}
