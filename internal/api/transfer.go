package api

import (
	"fmt"
	"net/http"

	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

type (
	transferRequest struct {
		FromAccountID int64  `json:"from_account_id" validate:"required,min=1"`
		ToAccountID   int64  `json:"to_account_id" validate:"required,min=1"`
		Amount        int64  `json:"amount" validate:"required,gt=0"`
		Currency      string `json:"currency" validate:"required,currency"`
	}
)

func (server *Server) createTransfer(ectx echo.Context) (err error) {
	var result db.TransferTxResult

	req := &transferRequest{}

	if err = ectx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = ectx.Validate(req); err != nil {
		return err
	}

	if ok, err := server.isSameCurrency(ectx, req.FromAccountID, req.Currency); !ok {
		return err
	}

	if ok, err := server.isSameCurrency(ectx, req.ToAccountID, req.Currency); !ok {
		return err
	}

	arg := db.CreateTransferParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	if result, err = server.store.TransferTx(ectx.Request().Context(), arg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, result)
}

func (server *Server) isSameCurrency(ectx echo.Context, accountID int64, currency string) (isOk bool, err error) {
	var account db.Account

	if account, err = server.store.GetAccount(ectx.Request().Context(), accountID); err != nil {
		if err == pgx.ErrNoRows {
			err = echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			err = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return
	}

	if account.Currency != currency {
		err = echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Sprintf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency),
		)

		return
	}

	isOk = true

	return
}
