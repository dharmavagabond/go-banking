package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	_ "github.com/dharmavagabond/simple-bank/doc/statik"
	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/http/grpc"
	"github.com/dharmavagabond/simple-bank/internal/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	var eg errgroup.Group

	store := db.NewStore()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	eg.Go(func() (err error) {
		if err = runGatewayServer(store); err != nil {
			err = fmt.Errorf("Gateway Server: %w", err)
		}

		return err
	})
	eg.Go(func() (err error) {
		if err = runGrpcServer(store); err != nil {
			err = fmt.Errorf("gRPC server: %w", err)
		}

		return err
	})

	if err := eg.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Err")
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
		statikFs http.FileSystem
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

	if statikFs, err = fs.New(); err != nil {
		return err
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))

	mux.Handle("/", grpcMux)
	mux.Handle("/swagger/", swaggerHandler)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return err
	}

	log.Info().Msgf("Listening HTTP gateway at %s", listener.Addr().String())

	return http.Serve(listener, grpc.HttpLogger(mux))
}
