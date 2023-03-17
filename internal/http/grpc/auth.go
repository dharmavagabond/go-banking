package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dharmavagabond/simple-bank/internal/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (payload *token.Payload, err error) {
	var (
		values      []string
		authType    string
		accessToken string
	)

	if md, ok := metadata.FromIncomingContext(ctx); !ok {
		return nil, errors.New("missing metadata")
	} else {
		if values = md.Get(authorizationHeader); len(values) == 0 {
			return nil, errors.New("missing authorization header")
		}
	}

	if fields := strings.Fields(values[0]); len(fields) < 2 {
		return nil, errors.New("invalid authorization header format")
	} else {
		authType = strings.ToLower(fields[0])
		accessToken = fields[1]
	}

	if authType != authorizationBearer {
		return nil, errors.New("unsupported authorization type")
	}

	if payload, err = server.tokenMaker.VerifyToken(accessToken); err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	return payload, nil
}
