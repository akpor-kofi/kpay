package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/ports"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
)

type MerchantHandler struct {
	merchantRepo ports.MerchantRepository

	ctx context.Context
}

func NewMerchantHandler(ctx context.Context, repo1 ports.MerchantRepository) *MerchantHandler {
	return &MerchantHandler{
		merchantRepo: repo1,

		ctx: ctx,
	}
}

func (mh *MerchantHandler) Add(c *fiber.Ctx) error {
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

	// encrypt password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newMerchant.Password), 12)
	newMerchant.Password = string(hashedPassword)

	// generate secret
	r := make([]byte, 32)
	rand.Read(r)
	newMerchant.Secret = hex.EncodeToString(r)

	// add merchant to database
	m, err := mh.merchantRepo.Add(newMerchant)
	if err != nil {
		return err
	}

	return c.JSON(m)
}

func validateMerchant(merchant domain.Merchant) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(merchant)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
