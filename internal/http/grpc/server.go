package grpc

import (
	"net"
	"strconv"

	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/pb"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/dharmavagabond/simple-bank/internal/worker"
	"github.com/rs/zerolog/log"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type (
	Server struct {
		pb.UnimplementedSimpleBankServer
		store           db.Store
		tokenMaker      token.Maker
		taskDistributor worker.TaskDistributor
	}
)

func (server *Server) Start() error {
	var (
		listener net.Listener
		err      error
	)

	addr := net.JoinHostPort(config.App.Host, strconv.Itoa(config.App.GrpcPort))
	grpcLogger := ggrpc.UnaryInterceptor(gRpcLogger)
	rpcServer := ggrpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServer(rpcServer, server)
	reflection.Register(rpcServer)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return err
	}

	log.Info().Msgf("Listening gRPC at %s", listener.Addr().String())

	return rpcServer.Serve(listener)
}

func NewServer(store db.Store, taskDistributor worker.TaskDistributor) (server *Server, err error) {
	var tokenMaker token.Maker

	if tokenMaker, err = token.NewPasetoMaker(config.App.TokenSymmetricKey); err != nil {
		return nil, err
	}

	server = &Server{
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}
