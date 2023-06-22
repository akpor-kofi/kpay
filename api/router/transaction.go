package router

import (
	"context"
	"fraud-detect-system/api/handler"
	"fraud-detect-system/mailer"
	"fraud-detect-system/services/fraud_detector_srv"
	"fraud-detect-system/services/transaction_srv"
	"fraud-detect-system/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewTransactionRouter(ctx context.Context, router fiber.Router) {
	transactionStorage := storage.NewTransactionStorage(ctx, "transactions")
	accountStorage := storage.NewAccountStorage(ctx, "accounts")

	transactionService := transaction_srv.New(accountStorage, transactionStorage)
	fraudDetectorService := fraud_detector_srv.New(accountStorage, transactionStorage)

	m := mailer.New(&mailer.Config{
		Host:     "sandbox.smtp.mailtrap.io",
		Port:     "25",
		Username: "1c4aa645c836ad",
		Password: "12e09da5cc43e8",
	})

	th := handler.NewTransactionHandler(ctx, transactionStorage, accountStorage, *transactionService, *fraudDetectorService, m)

	router.Use(cors.New(cors.Config{
		AllowMethods: "GET,POST",
		AllowHeaders: "allow",
	}))

	router.Post("/initialize", th.InitializeTransaction)

	router.Post("/transactions", th.AddTransaction)

	//router.Post("/transactions/refactored")

	router.Get("/transactions", th.GetAll)

	router.Get("/hello", func(ctx *fiber.Ctx) error {
		return ctx.SendString("hello")
	})
}
