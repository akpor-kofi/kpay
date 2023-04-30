package cmd

//
//import (
//	"context"
//	"fraud-detect-system/api/router"
//	"github.com/gofiber/fiber/v2"
//	"github.com/gofiber/fiber/v2/middleware/compress"
//	"github.com/kamva/mgm/v3"
//	"go.mongodb.org/mongo-driver/mongo/options"
//	"log"
//)
//
//func main() {
//	//mongoUrl := os.Getenv("MONGO_URL")
//	log.Printf("print here first")
//
//	err := mgm.SetDefaultConfig(nil, "fraudis_dev", options.Client().ApplyURI("mongodb://localhost:27018"))
//	if err != nil {
//		panic(err)
//	}
//
//	// connect app
//	log.Printf("Fiber cold start")
//
//	var app *fiber.App
//	app = fiber.New()
//
//	api := app.Group("/v1", compress.New())
//
//	router.NewTransactionRouter(context.TODO(), api)
//	router.NewMerchantRouter(context.TODO(), api)
//
//	err = app.Listen(":3000")
//
//	if err != nil {
//		panic(err)
//	}
//}
