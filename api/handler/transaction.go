package handler

import (
	"context"
	"fmt"
	"fraud-detect-system/clustering"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/ports"
	"github.com/e-XpertSolutions/go-iforest/v2/iforest"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentJSON struct {
	Amount      float64 `json:"amount" validate:"required,number,min=0"`
	CardHolder  string  `json:"card_holder" validate:"required"`
	ExpiryMonth int     `json:"expiry_month" validate:"required"`
	ExpiryYear  int     `json:"expiry_year" validate:"required"`
	Cvc         string  `json:"cvc" validate:"required"`
	CreditCard  string  `json:"credit_card" validate:"required,credit_card"`
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

func NewTransactionHandler(ctx context.Context, repo1 ports.TransactionRepository, repo2 ports.AccountRepository) *TransactionHandler {
	return &TransactionHandler{
		transactionRepo: repo1,
		accountRepo:     repo2,

		ctx: ctx,
	}
}

func (th *TransactionHandler) Add(c *fiber.Ctx) error {
	payment := c.UserContext().Value("payment").(PaymentJSON)

	newTransaction := domain.Transaction{
		Amt:        payment.Amount,
		CreditCard: payment.CreditCard,
		//Ip:         c.GetReqHeaders()["X-Forwarded-For"],
		//UserAgent:  c.GetReqHeaders()["User-Agent"],
		Ip:        c.IP(),
		UserAgent: string(c.Context().UserAgent()),
		Merchant:  payment.Merchant,
		User: domain.User{
			Name:  payment.Name,
			Email: payment.Email,
			Phone: payment.Phone,
		},
	}

	_, err := th.transactionRepo.Add(newTransaction)
	if err != nil {
		return err
	}

	transactions, err := th.transactionRepo.GetAllWhereAccountIs(payment.CreditCard)
	if err != nil {
		return err
	}

	child := context.WithValue(c.UserContext(), "transactions", transactions)
	c.SetUserContext(child)

	return c.Next()

}

func (th *TransactionHandler) GetAccount(c *fiber.Ctx) error {
	var payment PaymentJSON
	if err := c.BodyParser(&payment); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	errors := validatePaymentJSON(payment)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)

	}

	account, err := th.accountRepo.Get(payment.CreditCard)

	// creates an account if it doesn't exist
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// create the account
			newAccount := domain.Account{
				Merchant:   payment.Merchant,
				IPAddress:  c.GetReqHeaders()["X-Forwarded-For"],
				CardHolder: payment.CardHolder,
				CardNum:    payment.CreditCard,
			}

			newAccount, err = th.accountRepo.Create(newAccount)
			if err != nil {
				return err
			}

			account = newAccount
		} else {
			panic(err)
		}
	}

	parent := c.UserContext()
	accountCtx := context.WithValue(parent, "account", account)
	accountCtx = context.WithValue(accountCtx, "payment", payment)

	c.SetUserContext(accountCtx)
	return c.Next()
}

func (th *TransactionHandler) ClusterAccountTransactions(c *fiber.Ctx) error {
	account := c.UserContext().Value("account").(domain.Account)

	// --> Get clusters
	err := clustering.ClusterTransactions(account, th.transactionRepo, th.accountRepo)
	if err != nil {
		return err
	}

	return c.Next()
}

func (th *TransactionHandler) TrainAccountModel(c *fiber.Ctx) error {
	account := c.UserContext().Value("account").(domain.Account)

	//--> train the model by feeding it observations of transactions
	transactions := c.UserContext().Value("transactions").([]domain.Transaction)

	var amts []float64
	var obs []int

	amts = transform(transactions[len(transactions)-10:])

	// TODO: should make a function that pads the array if len(amts) < 10

	obs = classifyHabits(account, amts)
	account.Fit(obs)
	// --> save the trained model
	err := th.accountRepo.Save(&account)
	if err != nil {
		return err
	}

	child := context.WithValue(c.UserContext(), "obs", obs)
	c.SetUserContext(child)

	return c.Next()
}

// TrainIForest background job
func (th *TransactionHandler) TrainIForest(c *fiber.Ctx) error {
	//--> train the model by feeding it observations of transactions
	transactions := c.UserContext().Value("transactions").([]domain.Transaction)

	fmt.Println(len(transactions))

	inputData := th.preprocess(transactions)

	fmt.Println(inputData)

	treesNumber := 100
	subsampleSize := 256
	outliersRatio := 0.04 //0.01
	routinesNumber := 10

	//model initialization
	forest := iforest.NewForest(treesNumber, subsampleSize, outliersRatio)

	//training stage - creating trees
	forest.Train(inputData)

	err := forest.TestParallel(inputData, routinesNumber)
	if err != nil {
		panic(err)
	}

	threshold := forest.AnomalyBound
	anomalyScores := forest.AnomalyScores
	labelsTest := forest.Labels

	labelMap := map[int]int{}

	for i, label := range labelsTest {
		labelMap[i] = label
	}

	return c.JSON(fiber.Map{
		"threshold":      threshold,
		"anomaly_scores": anomalyScores,
		"labels_test":    labelMap,
	})
}

func (th *TransactionHandler) DetectFraud(c *fiber.Ctx) error {
	account := c.UserContext().Value("account").(domain.Account)
	obs := c.UserContext().Value("obs").([]int)

	// --> detect fraud
	//if isFraud, _ := account.DetectFraud(obs); isFraud {
	//	// --> probably send an info to the merchant to investigate this account
	//
	//	// --> set the transaction as fraudulent
	//}

	isFraud, fraudProb := account.DetectFraud(obs)

	// 1) update transaction risk score
	transactions, err := th.transactionRepo.GetAllWhereAccountIs(account.CardNum)
	if err != nil {
		return err
	}

	// 2) get last transaction and update risk_score and is_fraud
	currTransaction := transactions[len(transactions)-1]

	currTransaction.RiskScore = fraudProb
	currTransaction.IsFraud = isFraud

	err = th.transactionRepo.Save(&currTransaction)
	if err != nil {
		return err
	}

	// --> Process purchase

	return c.Status(200).JSON(fiber.Map{
		"is_fraud":   isFraud,
		"fraud_prob": fraudProb,
	})
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

func (th *TransactionHandler) preprocess(transactions []domain.Transaction) [][]float64 {
	inputData := make([][]float64, len(transactions))
	allTransactions, _ := th.transactionRepo.GetAll()

	// merchant
	var evaluateMerchant = func(transaction domain.Transaction, pos int) float64 {
		for _, tx := range transactions[:pos] {
			if transaction.Merchant == tx.Merchant {
				return 1
			}
		}

		for _, tx := range allTransactions {
			if tx.DefaultModel.CreatedAt.After(transaction.DefaultModel.CreatedAt) {
				break
			}
			if transaction.Merchant == tx.Merchant {
				return 0
			}
		}

		return -1
	}

	var evaluateIp = func(transaction domain.Transaction, pos int) float64 {
		for _, tx := range transactions[:pos] {
			if transaction.Ip == tx.Ip {
				return 1
			}
		}

		for _, tx := range allTransactions {
			if tx.DefaultModel.CreatedAt.After(transaction.DefaultModel.CreatedAt) {
				break
			}
			if transaction.Ip == tx.Ip {
				return 0
			}
		}

		return -1
	}

	var evaluateEmail = func(transaction domain.Transaction, pos int) float64 {
		for _, tx := range transactions[:pos] {
			if transaction.Email == tx.Email {
				return 1
			}
		}

		for _, tx := range allTransactions {
			if tx.DefaultModel.CreatedAt.After(transaction.DefaultModel.CreatedAt) {
				break
			}
			if transaction.Email == tx.Email {
				return 0
			}
		}

		return -1
	}

	var evaluateName = func(transaction domain.Transaction, pos int) float64 {
		for _, tx := range transactions[:pos] {
			if transaction.Name == tx.Name {
				return 1
			}
		}

		for _, tx := range allTransactions {
			if tx.DefaultModel.CreatedAt.After(transaction.DefaultModel.CreatedAt) {
				break
			}
			if transaction.Name == tx.Name {
				return 0
			}
		}

		return -1
	}

	var evaluateUserAgent = func(transaction domain.Transaction, pos int) float64 {
		for _, tx := range transactions[:pos] {
			if transaction.UserAgent == tx.UserAgent {
				return 1
			}
		}

		return -1
	}

	var evaluatePhone = func(transaction domain.Transaction, pos int) float64 {
		for _, tx := range transactions[:pos] {
			if transaction.Phone == tx.Phone {
				return 1
			}
		}

		for _, tx := range allTransactions {
			if tx.DefaultModel.CreatedAt.After(transaction.DefaultModel.CreatedAt) {
				break
			}
			if transaction.Phone == tx.Phone {
				return 0
			}
		}

		return -1
	}

	for i, transaction := range transactions {
		fmt.Println(transaction.Email)
		var rowData []float64
		rowData = append(rowData, evaluateMerchant(transaction, i))
		rowData = append(rowData, evaluateEmail(transaction, i))
		rowData = append(rowData, evaluateIp(transaction, i))
		rowData = append(rowData, evaluateUserAgent(transaction, i))
		rowData = append(rowData, evaluateName(transaction, i))
		rowData = append(rowData, evaluatePhone(transaction, i))

		inputData[i] = rowData
	}
	// merchant
	// email
	// ip address
	// user agent
	// name
	// phone

	// 1 - if it's the same one the user has been using
	// 0 - if it's not the same but is in the kofipay network in some other transaction
	// -1 - if it's not the same and if it's not in the network at all

	return inputData
}

func transform(txs []domain.Transaction) (amounts []float64) {
	for _, tx := range txs {
		amounts = append(amounts, tx.Amt)
	}
	return
}

func classifyHabits(account domain.Account, amounts []float64) (obs []int) {
	for _, amount := range amounts {
		o := account.ClassifyTransaction(amount)
		obs = append(obs, o)
	}

	return
}
