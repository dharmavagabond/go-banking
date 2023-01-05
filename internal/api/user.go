package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

type (
	createUserRequest struct {
		Username string `json:"username" validate:"required,alphanum"`
		Password string `json:"password" validate:"required,min=10"`
		FullName string `json:"full_name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
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
		AccessToken string       `json:"access_token"`
		User        userResponse `json:"user"`
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
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
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
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return echo.NewHTTPError(http.StatusForbidden, err.Error())
			}
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, newUserResponse(user))
}

func (server *Server) loginUser(ectx echo.Context) (err error) {
	var (
		accessToken string
		user        db.User
		ok          bool
		req         = &loginUserRequest{}
	)

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	if user, err = server.store.GetUser(ectx.Request().Context(), req.Username); err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if ok, err = argon2id.ComparePasswordAndHash(req.Password, user.HashedPassword); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Contraseña incorrecta.")
	}

	if accessToken, err = server.tokenMaker.CreateToken(user.Username, config.App.AccessTokenDuration); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}

	return ectx.JSON(http.StatusOK, res)
}
