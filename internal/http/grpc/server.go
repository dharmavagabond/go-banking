package grpc

import (
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	pb "github.com/dharmavagabond/simple-bank/internal/pb/user/v1"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/dharmavagabond/simple-bank/internal/worker"
)

type (
	Server struct {
		pb.UnimplementedSimpleBankServiceServer
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
	grpcLogger := ggrpc.UnaryInterceptor(gRPCLogger)
	rpcServer := ggrpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServiceServer(rpcServer, server)
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
