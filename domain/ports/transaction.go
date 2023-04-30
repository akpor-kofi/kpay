package ports

import "fraud-detect-system/domain"

type TransactionRepository interface {
	Add(newTransaction domain.Transaction) (domain.Transaction, error)
	GetAllWhereAccountIs(creditCard string) ([]domain.Transaction, error)
	GetAll() ([]domain.Transaction, error)
	Save(transaction *domain.Transaction) error
}
