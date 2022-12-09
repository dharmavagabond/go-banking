package api

import (
	"errors"
	"net/http"
	"time"

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
	createUserResponse struct {
		Username          string    `json:"username"`
		FullName          string    `json:"full_name"`
		Email             string    `json:"email"`
		PasswordChangedAt time.Time `json:"password_changed_at"`
		CreatedAt         time.Time `json:"created_at"`
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

	crtUserRes := createUserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}

	return ectx.JSON(http.StatusOK, crtUserRes)
}
