package api

import (
	"net/http"

	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

type (
	createAccountRequest struct {
		Owner    string `json:"owner" validate:"required"`
		Currency string `json:"currency" validate:"required,oneof=USD EUR"`
	}

	getAccountRequest struct {
		ID int64 `param:"id" validate:"required,min=1"`
	}

	listAccountRequest struct {
		PageID   int32 `query:"page_id" validate:"required,min=1"`
		PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
	}
)

func (server *Server) createAccount(ectx echo.Context) (err error) {
	var account db.Account
	req := &createAccountRequest{}

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
	}

	if account, err = server.store.CreateAccount(ectx.Request().Context(), arg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, account)
}

func (server *Server) getAccount(ectx echo.Context) (err error) {
	var account db.Account
	req := &getAccountRequest{}

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	if account, err = server.store.GetAccount(ectx.Request().Context(), req.ID); err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, account)
}

func (server *Server) listAccounts(ectx echo.Context) (err error) {
	var accounts []db.Account
	req := &listAccountRequest{
		PageID:   1,
		PageSize: 5,
	}

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	arg := db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	if accounts, err = server.store.ListAccounts(ectx.Request().Context(), arg); err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, accounts)
}
