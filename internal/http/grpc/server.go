package grpc

import (
	"log"
	"net"
	"strconv"

	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/pb"
	"github.com/dharmavagabond/simple-bank/internal/token"
	rpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type (
	Server struct {
		pb.UnimplementedSimpleBankServer
		store      db.Store
		tokenMaker token.Maker
	}
)

func (server *Server) Start() error {
	var (
		listener net.Listener
		err      error
	)

	rpcServer := rpc.NewServer()
	addr := net.JoinHostPort(config.App.Host, strconv.Itoa(config.App.GrpcPort))

	pb.RegisterSimpleBankServer(rpcServer, server)
	reflection.Register(rpcServer)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return err
	}

	log.Printf("Listening gRPC at %s", listener.Addr().String())

	return rpcServer.Serve(listener)
}

func NewServer(store db.Store) (server *Server, err error) {
	var tokenMaker token.Maker

	if tokenMaker, err = token.NewPasetoMaker(config.App.TokenSymmetricKey); err != nil {
		return nil, err
	}

	server = &Server{
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
