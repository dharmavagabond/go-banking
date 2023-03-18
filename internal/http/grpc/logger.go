package grpc

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func gRpcLogger(ctx context.Context, req interface{}, info *ggrpc.UnaryServerInfo, handler ggrpc.UnaryHandler) (interface{}, error) {
	var logger *zerolog.Event

	startTime := time.Now()
	res, err := handler(ctx, req)
	duration := time.Since(startTime)
	statusCode := codes.Unknown

	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	if err != nil {
		logger = log.Error().Err(err)
	} else {
		logger = log.Info()
	}

	logger.
		Str("protocol", "gRPC").
		Str("method", info.FullMethod).
		Dict(
			"status",
			zerolog.Dict().
				Int("code", int(statusCode)).
				Str("text", statusCode.String()),
		).
		Dur("duration", duration).
		Msg("received a gRPC request")

	return res, err
}
