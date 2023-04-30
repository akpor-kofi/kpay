package router

import (
	"context"
	"fraud-detect-system/api/handler"
	"fraud-detect-system/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewTransactionRouter(ctx context.Context, router fiber.Router) {
	transactionStorage := storage.NewTransactionStorage(ctx, "transactions")
	accountStorage := storage.NewAccountStorage(ctx, "accounts")

	th := handler.NewTransactionHandler(ctx, transactionStorage, accountStorage)

	router.Use(cors.New())

	router.Post("/transactions", th.GetAccount, th.Add, th.ClusterAccountTransactions, th.TrainAccountModel, th.TrainIForest, th.DetectFraud)

	router.Get("/transactions", th.GetAll)

	router.Get("/hello", func(ctx *fiber.Ctx) error {
		return ctx.SendString("hello")
	})

}
