package token

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

type Session struct {
	Email        string    `json:"email"`
	RefreshToken string    `json:"refresh_token"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiredAt    time.Time `json:"expired_at"`
}

func SessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

func NewSession(c *fiber.Ctx, p *Payload, refreshToken string) *Session {
	return &Session{
		Email:        p.Email,
		RefreshToken: refreshToken,
		//ClientIp:     c.IP(),
		ClientIp:  c.GetReqHeaders()["X-Forwarded-For"],
		IsBlocked: false,
		UserAgent: fiber.AcquireAgent().Name,
		CreatedAt: time.Now(),
		ExpiredAt: p.ExpiredAt,
	}
}
