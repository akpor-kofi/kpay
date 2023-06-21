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
	transactionRepo := storage.NewTransactionStorage(ctx, "transactions")

	mh := handler.NewMerchantHandler(ctx, merchantRepo, transactionRepo)
	ah := handler.NewAuthHandler(ctx, merchantRepo)

	router.Use(cors.New())

	authRoute := router.Group("/")

	authRoute.Post("/login", ah.Login)

	authRoute.Get("/renew-access-token", ah.RenewAccessToken)
	authRoute.Post("/signup", ah.Signup)

	router.Post("/merchants", mh.Add)

	router.Get("/merchants/:id/transactions", ah.BasicAuth(), ah.RequireAuth, mh.GetMerchantTransactions)

	router.Get("/merchants/:id/overview", ah.BasicAuth(), ah.RequireAuth, mh.GetOverview)

	router.Get("/merchants/:id/related", ah.BasicAuth(), ah.RequireAuth, mh.GetRelatedPayments)
}
