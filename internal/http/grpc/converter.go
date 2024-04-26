package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	pb "github.com/dharmavagabond/simple-bank/internal/pb/user/v1"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
		CreatedAt:         timestamppb.New(user.CreatedAt.Time),
	}
}
