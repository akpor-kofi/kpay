package router

import (
	"context"
	"fraud-detect-system/api/handler"
	"fraud-detect-system/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewMerchantRouter(ctx context.Context, router fiber.Router) {
	merchantRepo := storage.NewMerchantStorage(ctx, "merchants")

	mh := handler.NewMerchantHandler(ctx, merchantRepo)

	router.Use(cors.New())

	router.Post("/merchants", mh.Add)

}
