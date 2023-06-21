package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/ports"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"math"
	"math/rand"
	"time"
)

type MerchantHandler struct {
	merchantRepo    ports.MerchantRepository
	transactionRepo ports.TransactionRepository

	ctx context.Context
}

func NewMerchantHandler(ctx context.Context, repo1 ports.MerchantRepository, repo2 ports.TransactionRepository) *MerchantHandler {
	return &MerchantHandler{
		merchantRepo:    repo1,
		transactionRepo: repo2,

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

	// generate secret
	r := make([]byte, 32)
	rand.Read(r)
	newMerchant.Secret = hex.EncodeToString(r)

	// add merchant to database
	m, err := mh.merchantRepo.Add(&newMerchant)
	if err != nil {
		return err
	}

	return c.JSON(m)
}

func (mh *MerchantHandler) GetMerchantTransactions(c *fiber.Ctx) error {
	merchantId := c.Params("id")

	transactions, err := mh.transactionRepo.GetAllWhereMerchantIs(merchantId)
	if err != nil {
		return err
	}

	return c.JSON(transactions)
}

func (mh *MerchantHandler) GetRelatedPayments(c *fiber.Ctx) error {
	// get keyword (either ip-address or state or transaction ip or date, or )
	prefix := c.Query("prefix")
	merchantId := c.Params("id")

	// store them in an array then send
	relatedTransactions, err := mh.transactionRepo.GetRelatedPayments(merchantId, prefix)

	if err != nil {
		return err
	}

	return c.JSON(relatedTransactions)
}

type MerchantOverview struct {
	Attempted         int                  `json:"attempted"`
	NotificationCount int                  `json:"notification_count"`
	Block             int                  `json:"block"`
	BlockRate         float64              `json:"block_rate"`
	VolumeBlocked     float64              `json:"volume_blocked"`
	RecentTransaction []domain.Transaction `json:"recent_transactions"`
	ChartData         ChartData            `json:"chart_data"`
}

type ChartData struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
}

func (mh *MerchantHandler) GetOverview(c *fiber.Ctx) error {
	// attempted --> number of transactions
	// block --> number of transactions where is_fraud is true
	// block_rate --> block / attempted * 100
	// volume_blocked --> sum total amt field of transactions where is_fraud is true
	// chart_data --> {labels: string[], datasets}

	merchantId := c.Params("id")

	// TODO: ADD FILTER BY DURATION

	transactions, err := mh.transactionRepo.GetAllWhereMerchantIs(merchantId)

	merchant, err := mh.merchantRepo.GetById(merchantId)

	if err != nil {
		return nil
	}

	attempted := len(transactions)
	block := 0
	volumeBlocked := 0.0

	for _, transaction := range transactions {
		if transaction.IsFraud {
			block++
			volumeBlocked += transaction.Amt
		}
	}

	blockRate := float64(block) / float64(attempted) * 100

	if math.IsNaN(blockRate) {
		blockRate = 0
	}

	// recent transactions
	var recentTransactions []domain.Transaction
	limit := 6
	count := 0

	for i := len(transactions) - 1; i >= 0; i-- {
		if count >= limit {
			break
		}
		recentTransactions = append(recentTransactions, transactions[i])
		count++
	}

	// chart data
	chartData := mh._getChartData(merchant)

	notificationCount := mh.transactionRepo.GetNewTransactions(merchantId, merchant.LastLoggedIn)

	overview := MerchantOverview{
		Attempted:         attempted,
		Block:             block,
		BlockRate:         blockRate,
		VolumeBlocked:     volumeBlocked,
		NotificationCount: notificationCount,
		RecentTransaction: recentTransactions,
		ChartData:         chartData,
	}

	fmt.Println(overview)

	merchant.LastLoggedIn = time.Now().UTC()
	err = mh.merchantRepo.Save(&merchant)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(overview)
}

func (mh *MerchantHandler) _getChartData(merchant domain.Merchant) ChartData {
	// Get the last 20 days starting from today
	today := time.Now()
	labels := make([]string, 0)
	dataFraudulent := make([]float64, 0)
	dataNonFraudulent := make([]float64, 0)

	for i := 19; i >= 0; i-- {
		// Calculate the date for each day
		date := today.AddDate(0, 0, -i)
		label := date.Format("January 2") // Format the date as "Month Day"
		labels = append(labels, label)

		// Fetch the total fraudulent and non-fraudulent transactions for that day
		totalFraudulent := mh._fetchTotalFraudulentTransactions(date, merchant.ID.Hex())
		totalNonFraudulent := mh._fetchTotalNonFraudulentTransactions(date, merchant.ID.Hex())

		dataFraudulent = append(dataFraudulent, totalFraudulent)
		dataNonFraudulent = append(dataNonFraudulent, totalNonFraudulent)
	}

	// Create the chart data struct
	chartData := ChartData{
		Labels: labels,
		Datasets: []Dataset{
			{
				Label: "Valid Transactions",
				Data:  dataNonFraudulent,
			},
			{
				Label: "Fraudulent Transactions",
				Data:  dataFraudulent,
			},
		},
	}

	return chartData
}

func (mh *MerchantHandler) _fetchTotalNonFraudulentTransactions(date time.Time, merchant string) float64 {
	nonFraudulentTransactions, _ := mh.transactionRepo.FetchTransactions(date, false, merchant)

	return float64(len(nonFraudulentTransactions))
}

func (mh *MerchantHandler) _fetchTotalFraudulentTransactions(date time.Time, merchant string) float64 {
	fraudulentTransactions, _ := mh.transactionRepo.FetchTransactions(date, true, merchant)

	return float64(len(fraudulentTransactions))
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
