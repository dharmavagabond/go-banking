package token

import (
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	IssuedAt  time.Time `json:"iat"`
	ExpiredAt time.Time `json:"exp"`
	Username  string    `json:"username"`
	ID        uuid.UUID `json:"id"`
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ERR_EXPIRED_TOKEN
	}

	return nil
}

func NewPayload(username string, duration time.Duration) (payload *Payload, err error) {
	payload = &Payload{
		ID:        uuid.New(),
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return
}
