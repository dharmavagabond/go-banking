package grpc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/pb"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/dharmavagabond/simple-bank/internal/valid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
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

	if violations := validateCreateUserRequest(req); len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

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

	if violations := validateLoginUserRequest(req); len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

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

	mtdt := server.extractMetadata(ctx)

	if session, err = server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshTokenPayload.ID,
		Username:     req.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
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

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (res *pb.UpdateUserResponse, err error) {
	var (
		authPayload      *token.Payload
		user             db.User
		hashPassword     string
		isPasswordHashed bool
	)

	if authPayload, err = server.authorizeUser(ctx); err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validateUpdateUserRequest(req); len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

	if authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	if len(req.GetPassword()) > 0 {
		if hashPassword, err = argon2id.CreateHash(req.GetPassword(), argonParams); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash the password: %s", err.Error())
		}

		isPasswordHashed = true
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		HashedPassword: sql.NullString{
			String: hashPassword,
			Valid:  isPasswordHashed,
		},
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  len(req.GetFullName()) > 0,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  len(req.GetEmail()) > 0,
		},
		PasswordChangedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: isPasswordHashed,
		},
	}

	if user, err = server.store.UpdateUser(ctx, arg); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not Found: %s", err.Error())
		}

		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err.Error())
	}

	res = &pb.UpdateUserResponse{
		User: convertUser(user),
	}

	return res, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := valid.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := valid.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := valid.ValidateFullname(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := valid.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := valid.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := valid.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return violations
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := valid.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if len(req.GetPassword()) > 0 {
		if err := valid.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if len(req.GetFullName()) > 0 {
		if err := valid.ValidateFullname(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if len(req.GetEmail()) > 0 {
		if err := valid.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return violations
}
