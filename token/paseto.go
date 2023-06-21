package token

import (
	"errors"
	"github.com/o1egl/paseto"
	"time"
)

var pasetoMaker = paseto.NewV2()
var symmetricKey = []byte("YELLOW SUBMARINE, BLACK WIZARDRY")

const AccessTokenDuration = 20000 * time.Minute
const RefreshTokenDuration = 168 * time.Hour

var ErrInvalidToken = errors.New("invalid token")

func CreateAccessToken(email string, merchantId string) (string, *Payload, error) {
	payload := NewPayload(email, merchantId, AccessTokenDuration, "access")

	token, err := pasetoMaker.Encrypt(symmetricKey, payload, nil)
	if err != nil {
		return "", &Payload{}, err
	}

	return token, payload, nil
}

func CreateRefreshToken(email string, merchantId string) (string, *Payload, error) {
	payload := NewPayload(email, merchantId, RefreshTokenDuration, "refresh")

	token, err := pasetoMaker.Encrypt(symmetricKey, payload, nil)
	if err != nil {
		return "", &Payload{}, err
	}

	return token, payload, nil
}

func VerifyAccessToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := pasetoMaker.Decrypt(token, symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if payload.Type != "access" {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func VerifyRefreshToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := pasetoMaker.Decrypt(token, symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if payload.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
