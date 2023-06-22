package handler

import (
	"context"
	"fmt"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/ports"
	"fraud-detect-system/services/feature_extraction_srv"
	"fraud-detect-system/services/fraud_detector_srv"
	"fraud-detect-system/services/transaction_srv"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type PaymentJSON struct {
	Amount      float64 `json:"amount" validate:"required,number,min=0"`
	CardHolder  string  `json:"card_holder" validate:"required"`
	ExpiryMonth int     `json:"expiry_month" validate:"required"`
	ExpiryYear  int     `json:"expiry_year" validate:"required"`
	Cvc         string  `json:"cvc" validate:"required"`
	CreditCard  string  `json:"credit_card" validate:"required"`
	IPAddress   string  `json:"ip_address"`
	domain.User
	Merchant string `json:"merchant" validate:"required"`
}

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

type TransactionHandler struct {
	transactionRepo ports.TransactionRepository
	accountRepo     ports.AccountRepository

	transactionsService  transaction_srv.TransactionService
	fraudDetectorService fraud_detector_srv.FraudDetectorService

	mailer ports.IMailer

	ctx context.Context
}

var validate = validator.New()

func validatePaymentJSON(payment PaymentJSON) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(payment)
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

func NewTransactionHandler(ctx context.Context, repo1 ports.TransactionRepository, repo2 ports.AccountRepository, transactionsService transaction_srv.TransactionService, fraudDetectorService fraud_detector_srv.FraudDetectorService, mailer ports.IMailer) *TransactionHandler {
	return &TransactionHandler{
		transactionRepo: repo1,
		accountRepo:     repo2,

		transactionsService:  transactionsService,
		fraudDetectorService: fraudDetectorService,

		mailer: mailer,

		ctx: ctx,
	}
}

func (th *TransactionHandler) AddTransaction(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
	}()

	fmt.Println(c.GetReqHeaders()["allow"])

	var payment PaymentJSON
	if err := c.BodyParser(&payment); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	errors := validatePaymentJSON(payment)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	account, err := th.transactionsService.GetAccountByCreditCard(payment.CreditCard)

	// creates an account if it doesn't exist
	if err != nil {
		if err == transaction_srv.ErrNoAccountFound {
			// create the account
			newAccount := domain.Account{
				CardHolder: payment.CardHolder,
				CardNum:    payment.CreditCard,
			}

			newAccount, _ = th.transactionsService.CreateAccount(newAccount)

			account = newAccount
		} else {
			panic(err)
		}
	}

	// add the transaction
	//TODO: verify merchant

	newTransaction := domain.Transaction{
		Amt:        payment.Amount,
		CreditCard: payment.CreditCard,
		//Ip:         strings.Split(c.GetReqHeaders()["X-Forwarded-For"], ",")[0],
		Ip:        c.GetReqHeaders()["X-Forwarded-For"],
		UserAgent: c.GetReqHeaders()["User-Agent"],
		//Ip:        c.IP(),
		//UserAgent: string(c.Context().UserAgent()),
		Merchant: payment.Merchant,
		User: domain.User{
			Email: payment.Email,
		},
	}

	newTransaction, err = th.transactionsService.Create(newTransaction)
	if err != nil {
		return err
	}

	transactions, err := th.transactionsService.GetAllValidByCreditCard(payment.CreditCard)
	if err != nil {
		return err
	}

	features := th.extractFeatures(newTransaction)
	xgbPred, err := th.fraudDetectorService.DetectXGB(features)
	if err != nil {
		return err
	}

	if len(transactions) < 20 {
		if xgbPred.Probability[1] >= 0.2 {
			newTransaction.IsFraud = true
		}

		newTransaction.RiskScore = xgbPred.Probability[1]

		th.transactionRepo.Save(&newTransaction)
		return c.JSON(xgbPred)
	}

	hmmPred, err := th.fraudDetectorService.DetectHMM(&account)

	finalPred := th.fraudDetectorService.Ensemble(hmmPred, xgbPred)

	if finalPred.Likelihood >= 0.7 || finalPred.Prob >= 0.2 || finalPred.IsHmmFraud {
		// probably fraud

		newTransaction.IsFraud = true

	}

	newTransaction.RiskScore = finalPred.Prob

	th.transactionRepo.Save(&newTransaction)

	//	_ = th.mailer.Send(domain.Mail{
	//		To:      "akporkofi11@gmail.com",
	//		Subject: "answer me",
	//		Message: `To: kate.doe@example.com
	//
	//From: john.doe@your.domain
	//
	//Subject: Why aren’t you using Mailtrap yet?
	//
	//Here’s the space for your great sales pitch`,
	//	})
	//if err != nil {
	//	return err
	//}

	return c.Status(200).JSON(th.fraudDetectorService.Ensemble(hmmPred, xgbPred))
}

func (th *TransactionHandler) GetAll(c *fiber.Ctx) error {
	fmt.Println("ip address")
	fmt.Println(c.IP() + " with ctx")
	fmt.Println(c.GetReqHeaders()["X-Forwarded-For"])

	transactions, err := th.transactionRepo.GetAll()
	if err != nil {
		return fiber.NewError(0, err.Error())
	}

	return c.JSON(transactions)
}

func (th *TransactionHandler) extractFeatures(currentTransaction domain.Transaction) fraud_detector_srv.XGBFeatures {
	var xgb fraud_detector_srv.XGBFeatures

	extractionService := feature_extraction_srv.New(th.transactionRepo)

	xgb.Hour = extractionService.ExtractHour(currentTransaction)
	xgb.TravelSpeed = extractionService.ExtractTravelSpeed(currentTransaction)
	xgb.LastDayTransactionCount, xgb.LastDayFraudTransactionCount = extractionService.ExtractLast24HourCount(currentTransaction)
	xgb.Amount = currentTransaction.Amt
	xgb.CategoryIndex = 0

	return xgb
}

type Initiator struct {
	SecretKey       string  `json:"secret_key"`
	Email           string  `json:"email"` // or just username
	Amount          float64 `json:"amount"`
	ProductMetadata string  `json:"product_metadata"` // json blob of product metadata
	MerchantId      string  `json:"merchant_id"`
}

func (th *TransactionHandler) InitializeTransaction(c *fiber.Ctx) error {
	// parse json to Initiator struct

	// verify merchant id and secret_key

	// generate a reference that links to transaction id for a specified time

	// initialize transaction

	// return a link that the user can use to pay and reference id

	return nil
}
