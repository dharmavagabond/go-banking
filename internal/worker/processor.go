package worker

import (
	"context"
	"net"

	"github.com/dharmavagabond/simple-bank/internal/config"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func (proc *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendVerifyEmail, proc.ProcessTaskSendVerifyEmail)
	return proc.server.Start(mux)
}

func NewRedisTaskProcessor(store db.Store) TaskProcessor {
	rcopt := asynq.RedisClientOpt{
		Addr: net.JoinHostPort(config.Redis.Host, config.Redis.Port),
	}
	server := asynq.NewServer(
		rcopt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			Logger: NewWorkerLogger(),
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					log.Error().
						Err(err).
						Str("type", task.Type()).
						Bytes("payload", task.Payload()).
						Msg("[process-task]")
				},
			),
		})

	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}
