package api

import (
	"net/http"

	"github.com/MadAppGang/httplog"
	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type (
	Server struct {
		store  db.Store
		router *echo.Echo
	}

	customValidator struct {
		validator *validator.Validate
	}
)

func (cv *customValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}

func (server *Server) Start(address string) error {
	return server.router.Start(address)
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := echo.New()
	sbvalidator := validator.New()
	router.Debug = config.App.IsDev
	router.Validator = &customValidator{validator: sbvalidator}
	server.router = router

	if err := sbvalidator.RegisterValidation("currency", validCurrency); err != nil {
		router.Logger.Fatal(err)
	}

	router.Use(loggerMiddleware)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
	router.POST("/accounts", server.createAccount)
	router.POST("/transfers", server.createTransfer)

	return server
}

func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := httplog.Logger(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if err := next(c); err != nil {
				c.Error(err)
			}

			lrw, _ := rw.(httplog.ResponseWriter)
			lrw.Set(c.Response().Status, int(c.Response().Size))
		}))
		logger.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}
