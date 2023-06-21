package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/ports"
	"fraud-detect-system/token"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"math/rand"
)

type AuthHandler struct {
	merchantRepo ports.MerchantRepository
	store        map[string]token.Session

	ctx context.Context
}

func (h AuthHandler) Login(c *fiber.Ctx) error {
	b := &struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	}{}

	if err := c.BodyParser(b); err != nil {
		return err
	}

	merchant, err := h.merchantRepo.GetByEmail(b.Email)

	// check password match
	if err = merchant.PasswordMatches(b.Password); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	// create access and refresh token
	access, accessPayload, err := token.CreateAccessToken(b.Email, merchant.ID.Hex())
	if err != nil {
		return err
	}

	refresh, refreshPayload, err := token.CreateRefreshToken(b.Email, merchant.ID.Hex())
	if err != nil {
		return err
	}

	// create the session object to be stored
	tokenHash := token.NewSession(c, refreshPayload, refresh)

	// set the token hash struct to the session
	key := token.SessionKey(refreshPayload.ID)

	h.store[key] = *tokenHash

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"access_token":             access,
		"access_token_expired_at":  accessPayload.ExpiredAt,
		"refresh_token":            refresh,
		"refresh_token_expired_at": refreshPayload.ExpiredAt,
		"merchant":                 merchant,
		"session_id":               refreshPayload.ID,
	})

}

func (h AuthHandler) Signup(c *fiber.Ctx) error {
	var newMerchant domain.Merchant

	if err := c.BodyParser(&newMerchant); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// validate email and confirm password
	errors := validateMerchant(newMerchant)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	if newMerchant.Password != newMerchant.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("password don't match"),
		})
	}

	// generate secret
	r := make([]byte, 32)
	rand.Read(r)
	newMerchant.Secret = hex.EncodeToString(r)

	// add merchant to database
	m, err := h.merchantRepo.Add(&newMerchant)
	if err != nil {
		return err
	}

	///////////SET TOKENS////////////
	// create access and refresh token
	access, accessPayload, err := token.CreateAccessToken(m.Email, m.ID.Hex())
	if err != nil {
		return err
	}

	refresh, refreshPayload, err := token.CreateRefreshToken(m.Email, m.ID.Hex())
	if err != nil {
		return err
	}

	// create the session object to be stored
	tokenHash := token.NewSession(c, refreshPayload, refresh)

	// set the token hash struct to the session
	key := token.SessionKey(refreshPayload.ID)

	h.store[key] = *tokenHash

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"access_token":             access,
		"access_token_expired_at":  accessPayload.ExpiredAt,
		"refresh_token":            refresh,
		"refresh_token_expired_at": refreshPayload.ExpiredAt,
		"merchant":                 m,
		"session_id":               refreshPayload.ID,
	})
}

func (h AuthHandler) RequireAuth(c *fiber.Ctx) error {

	isAuth := c.Locals("isAuthenticated")

	if _, ok := isAuth.(bool); !ok {

		return c.Next()
	}

	// get access token from header or cookie
	accessToken := c.Get("access-token")

	// verify token
	payload, err := token.VerifyAccessToken(accessToken)
	switch err {
	case token.ErrInvalidToken: //TODO
		return err
	case token.ErrExpiredToken:
		return err
	}

	// TODO: not only auth that token is good. the resource you are accessing should only the one you want
	merchantId := c.Params("id")

	if merchantId != payload.Merchant {
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("cannot access resource"))
	}

	return c.Next()
}

func (h AuthHandler) BasicAuth() fiber.Handler {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{
			"kofi":  "password",
			"admin": "123456",
		},
		Unauthorized: func(c *fiber.Ctx) error {
			c.Locals("isAuthenticated", false)
			return c.Next()
		},
	})
}

func (h AuthHandler) RenewAccessToken(c *fiber.Ctx) error {
	b := c.Query("refresh")

	// verify the refresh token is of type refresh and has a valid payload
	refreshPayload, err := token.VerifyRefreshToken(b)
	if err != nil {
		return err
	}

	key := token.SessionKey(refreshPayload.ID)

	// get the refresh token session to run extra checks: blocked, ip, useragent, etc...
	tokenHash := h.store[key]

	if tokenHash.IsBlocked {
		return fiber.NewError(fiber.StatusBadRequest, "blocked session")
	}

	if tokenHash.Email != refreshPayload.Email {
		return fiber.NewError(fiber.StatusBadRequest, "incorrect session user")
	}

	// create new access token
	access, accessPayload, err := token.CreateAccessToken(refreshPayload.Email, refreshPayload.Merchant)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":            access,
		"access_token_expired_at": accessPayload.ExpiredAt,
	})
}

func NewAuthHandler(ctx context.Context, repo1 ports.MerchantRepository) *AuthHandler {
	return &AuthHandler{
		merchantRepo: repo1,
		store:        make(map[string]token.Session),

		ctx: ctx,
	}
}
