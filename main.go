package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/http/grpc"
	"github.com/dharmavagabond/simple-bank/internal/http/rest"
	"github.com/dharmavagabond/simple-bank/internal/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	var eg errgroup.Group

	logger := echo.New().Logger
	store := db.NewStore()

	eg.Go(func() error {
		return runGatewayServer(store)
	})
	eg.Go(func() error {
		return runGrpcServer(store)
	})

	if err := eg.Wait(); err != nil {
		logger.Fatal(err)
	}
}

func runGrpcServer(store db.Store) error {
	var (
		server *grpc.Server
		err    error
	)

	if server, err = grpc.NewServer(store); err != nil {
		return err
	}

	return server.Start()
}

func runGatewayServer(store db.Store) error {
	var (
		server   *grpc.Server
		listener net.Listener
		err      error
	)

	addr := net.JoinHostPort(config.App.Host, strconv.Itoa(config.App.HttpPort))
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if server, err = grpc.NewServer(store); err != nil {
		return err
	}

	if err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return err
	}

	log.Printf("Listening HTTP gateway at %s", listener.Addr().String())

	return http.Serve(listener, mux)
}

func runHttpServer(store db.Store) error {
	var (
		server *rest.Server
		err    error
	)

	if server, err = rest.NewServer(store); err != nil {
		return err
	}

	return server.Start()
}
