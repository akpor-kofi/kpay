package ports

import (
	"fraud-detect-system/domain"
	"time"
)

type TransactionRepository interface {
	Add(newTransaction domain.Transaction) (domain.Transaction, error)
	GetAllWhereAccountIs(creditCard string) ([]domain.Transaction, error)
	GetAllValidTransactionWhereAccountIs(creditCard string) ([]domain.Transaction, error)
	FetchTransactions(date time.Time, isFraud bool, merchant string) ([]domain.Transaction, error)
	GetAllWhereMerchantIs(merchant string) ([]domain.Transaction, error)
	GetAll() ([]domain.Transaction, error)
	Save(transaction *domain.Transaction) error
	GetRelatedPayments(merchantId string, prefix string) ([]domain.Transaction, error)
	GetNewTransactions(merchantId string, lastLoggedIn time.Time) int
}

type ReferenceRepository interface {
	Set(transactionId string) string
	Get(reference string) (string, error)
}
