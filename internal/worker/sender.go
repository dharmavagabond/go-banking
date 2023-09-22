package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

const TaskSendVerifyEmail = "task:send_verify_email"

func (distr *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) (err error) {
	var (
		bs   []byte
		task *asynq.Task
	)

	if bs, err = json.Marshal(payload); err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task = asynq.NewTask(TaskSendVerifyEmail, bs, opts...)

	if taskInfo, err := distr.client.EnqueueContext(ctx, task); err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	} else {
		log.Info().
			Str("id", taskInfo.ID).
			Str("type", taskInfo.Type).
			Bytes("payload", task.Payload()).
			Int("retries", taskInfo.MaxRetry).
			Str("queue", taskInfo.Queue).
			Msg("enqueue task")
	}

	return nil
}

func (proc *RedisTaskProcessor) ProcessTaskSendVerifyEmail(
	ctx context.Context,
	task *asynq.Task,
) (err error) {
	var (
		payload PayloadSendVerifyEmail
		user    db.User
	)
	if err = json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	if user, err = proc.store.GetUser(ctx, payload.Username); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user doesn't exists: %w", asynq.SkipRetry)
		}

		return fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: send email

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("processed task")

	return nil
}
