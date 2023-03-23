package worker

import (
	"context"
	"net"

	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor() TaskDistributor {
	rcopt := asynq.RedisClientOpt{
		Addr: net.JoinHostPort(config.Redis.Host, config.Redis.Port),
	}
	client := asynq.NewClient(rcopt)

	return &RedisTaskDistributor{
		client: client,
	}
}
