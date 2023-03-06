package grpc

import (
	"context"
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/pb"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var argonParams = &argon2id.Params{
	Memory:      128 * 1024,
	Iterations:  4,
	Parallelism: 4,
	SaltLength:  128,
	KeyLength:   128,
}

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (res *pb.CreateUserResponse, err error) {
	var (
		user         db.User
		hashPassword string
	)

	if hashPassword, err = argon2id.CreateHash(req.GetPassword(), argonParams); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash the password: %s", err.Error())
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}

	if user, err = server.store.CreateUser(ctx, arg); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err.Error())
			}
		}

		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err.Error())
	}

	res = &pb.CreateUserResponse{
		User: convertUser(user),
	}

	return res, nil
}

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (res *pb.LoginUserResponse, err error) {
	var (
		user                db.User
		ok                  bool
		session             db.Session
		accessToken         string
		accessTokenPayload  *token.Payload
		refreshToken        string
		refreshTokenPayload *token.Payload
	)

	if user, err = server.store.GetUser(ctx, req.GetUsername()); err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to find user: %s", err.Error())
	}

	if ok, err = argon2id.ComparePasswordAndHash(req.Password, user.HashedPassword); err != nil {
		return nil, status.Errorf(codes.Internal, "couldn't compare the password: %s", err.Error())
	} else if !ok {
		return nil, status.Error(codes.Internal, "incorrect password")
	}

	if accessToken, accessTokenPayload, err = server.tokenMaker.CreateToken(user.Username, config.App.AccessTokenDuration); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if refreshToken, refreshTokenPayload, err = server.tokenMaker.CreateToken(
		req.Username,
		config.App.RefreshTokenDuration,
	); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if session, err = server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshTokenPayload.ID,
		Username:     req.Username,
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
		ExpiresAt:    refreshTokenPayload.ExpiredAt,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res = &pb.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessTokenPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshTokenPayload.ExpiredAt),
		User:                  convertUser(user),
	}

	return res, nil
}
