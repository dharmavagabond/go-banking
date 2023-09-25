package grpc

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func init() {
	if config.App.IsDev {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func gRPCLogger(
	ctx context.Context,
	req interface{},
	info *ggrpc.UnaryServerInfo,
	handler ggrpc.UnaryHandler,
) (interface{}, error) {
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

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var logger *zerolog.Event

		startTime := time.Now()
		rec := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusContinue,
		}
		handler.ServeHTTP(rec, r)
		duration := time.Since(startTime)

		if rec.StatusCode <= 299 {
			logger = log.Info()
		} else {
			logger = log.Error().Bytes("body", rec.Body)
		}

		logger.
			Str("protocol", "http").
			Str("method", r.Method).
			Str("path", r.RequestURI).
			Dict(
				"status",
				zerolog.Dict().
					Int("code", rec.StatusCode).
					Str("text", http.StatusText(rec.StatusCode)),
			).
			Dur("duration", duration).
			Msg("received a HTTP request")
	})
}
