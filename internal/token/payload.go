package token

import (
	"time"

	"github.com/gofrs/uuid"
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"iat"`
	ExpiredAt time.Time `json:"exp"`
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ERR_EXPIRED_TOKEN
	}

	return nil
}

func NewPayload(username string, duration time.Duration) (payload *Payload, err error) {
	var tokenID uuid.UUID

	if tokenID, err = uuid.NewV4(); err != nil {
		return
	}

	payload = &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return
}
