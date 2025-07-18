package rest

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type (
	Server struct {
		store      db.Store
		tokenMaker token.Maker
		router     *echo.Echo
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

func (server *Server) Start() error {
	address := net.JoinHostPort("", strconv.Itoa(config.App.HTTPPort))
	return server.router.Start(address)
}

func NewServer(store db.Store) (server *Server, err error) {
	var tokenMaker token.Maker

	router := echo.New()

	if tokenMaker, err = token.NewPasetoMaker(config.App.TokenSymmetricKey); err != nil {
		return nil, fmt.Errorf("%w: %w", token.ERR_CANT_CREATE_TOKEN_MAKER, err)
	}

	server = &Server{
		store:      store,
		tokenMaker: tokenMaker,
	}
	sbvalidator := validator.New()
	router.Debug = config.App.IsDev
	router.Validator = &customValidator{validator: sbvalidator}
	server.router = router

	if err = sbvalidator.RegisterValidation("currency", validCurrency); err != nil {
		return nil, err
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	server.router.Use(loggerMiddleware)
	server.router.POST("/signin", server.loginUser)
	server.router.GET("/accounts", server.listAccounts, authMiddleware)
	server.router.GET("/accounts/:id", server.getAccount, authMiddleware)
	server.router.POST("/accounts", server.createAccount, authMiddleware)
	server.router.POST("/transfers", server.createTransfer, authMiddleware)
	server.router.POST("/users", server.createUser, authMiddleware)
	server.router.POST("/token/refresh", server.renewAccessToken)
}
