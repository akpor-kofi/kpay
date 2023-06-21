package token

import (
	"errors"
	"github.com/gofiber/fiber/v2/utils"
	"time"
)

var ErrExpiredToken = errors.New("token is expired")

type Payload struct {
	ID        string    `json:"id"`
	Merchant  string    `json:"merchant"`
	Email     string    `json:"email"`
	Type      string    `json:"type"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(email string, merchantId string, duration time.Duration, tokenType string) *Payload {
	payload := &Payload{
		ID:        utils.UUIDv4(),
		Merchant:  merchantId,
		Email:     email,
		Type:      tokenType,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
