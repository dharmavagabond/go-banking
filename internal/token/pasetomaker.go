package token

import (
	"encoding/json"
	"time"

	"aidanwoods.dev/go-paseto"
)

type PasetoMaker struct {
	pasetoToken  *paseto.Token
	symmetricKey paseto.V4SymmetricKey
}

func (maker *PasetoMaker) CreateToken(
	username string,
	duration time.Duration,
) (token string, payload *Payload, err error) {
	if payload, err = NewPayload(username, duration); err != nil {
		return
	}

	maker.pasetoToken.SetString("id", payload.ID.String())
	maker.pasetoToken.SetString("username", payload.Username)
	maker.pasetoToken.SetIssuedAt(payload.IssuedAt)
	maker.pasetoToken.SetExpiration(payload.ExpiredAt)
	token = maker.pasetoToken.V4Encrypt(maker.symmetricKey, nil)

	return
}

func (maker *PasetoMaker) VerifyToken(token string) (payload *Payload, err error) {
	var pasetoToken *paseto.Token
	parser := paseto.NewParser()

	if pasetoToken, err = parser.ParseV4Local(maker.symmetricKey, token, nil); err != nil {
		if err.Error() == "this token has expired" {
			return nil, ERR_EXPIRED_TOKEN
		}

		return nil, err
	}

	if err = json.Unmarshal(pasetoToken.ClaimsJSON(), &payload); err != nil {
		return nil, err
	}

	if err = payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewPasetoMaker(symmetricKey string) (maker Maker, err error) {
	var (
		pv4sk paseto.V4SymmetricKey
		token = paseto.NewToken()
	)

	if pv4sk, err = paseto.V4SymmetricKeyFromBytes([]byte(symmetricKey)); err != nil {
		return
	}

	maker = &PasetoMaker{
		pasetoToken:  &token,
		symmetricKey: pv4sk,
	}

	return
}
