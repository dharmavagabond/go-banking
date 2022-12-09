package api

import (
	"errors"
	"net/http"

	"github.com/alexedwards/argon2id"
	dberrors "github.com/dharmavagabond/simple-bank/internal/db"
	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/jackc/pgconn"
	"github.com/labstack/echo/v4"
)

type (
	createUserRequest struct {
		Username string `json:"username" validate:"required,alphanum"`
		Password string `json:"password" validate:"required,min=10"`
		FullName string `json:"full_name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
	}
)

var argonParams = &argon2id.Params{
	Memory:      128 * 1024,
	Iterations:  4,
	Parallelism: 4,
	SaltLength:  128,
	KeyLength:   128,
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
			case dberrors.ERRCODE_UNIQUE_VIOLATION:
				return echo.NewHTTPError(http.StatusForbidden, err.Error())
			}
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, user)
}
