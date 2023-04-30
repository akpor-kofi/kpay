package main

import (
	"context"
	"fmt"
	"fraud-detect-system/api/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/joho/godotenv"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("no .env file found")
		fmt.Println("production mode")
	}
	mode := os.Getenv("FIBER_ENV")
	mongoUrl := os.Getenv("MONGO_URL")
	railwayPort := os.Getenv("PORT")
	fmt.Println(railwayPort)

	err = mgm.SetDefaultConfig(nil, "fraudis_dev", options.Client().ApplyURI(mongoUrl))
	if err != nil {
		panic("cannot connect database")
	}

	// connect app
	log.Printf("Fiber cold start")

	var app *fiber.App
	app = fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello world")
	})

	api := app.Group("/v1", compress.New())

	router.NewTransactionRouter(context.TODO(), api)
	router.NewMerchantRouter(context.TODO(), api)

	if mode == "DEV" {
		err = app.Listen(":3000")
	} else {
		err = app.Listen("0.0.0.0:" + railwayPort)
	}

	if err != nil {
		panic(err)
	}
}
