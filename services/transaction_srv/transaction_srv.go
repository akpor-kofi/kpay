package transaction_srv

import (
	"errors"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/ports"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	low = iota
	avg
	high
)

var ErrNoAccountFound = errors.New("no account found, create one")

type TransactionService struct {
	transactionRepository ports.TransactionRepository
	accountRepository     ports.AccountRepository
}

func New(accountRepository ports.AccountRepository, transactionRepository ports.TransactionRepository) *TransactionService {
	return &TransactionService{
		transactionRepository: transactionRepository,
		accountRepository:     accountRepository,
	}
}

func (ts *TransactionService) SaveCluster(account *domain.Account, corePoints []float64) error {
	account.SpendingHabits[low] = corePoints[low] // TODO: make use of enum constant low, avg & high
	account.SpendingHabits[avg] = corePoints[avg]
	account.SpendingHabits[high] = corePoints[high]

	return ts.accountRepository.Save(account)
}

func (ts *TransactionService) GetAllValidByCreditCard(card string) ([]domain.Transaction, error) {
	return ts.transactionRepository.GetAllValidTransactionWhereAccountIs(card)
}

func (ts *TransactionService) GetAccountByCreditCard(card string) (domain.Account, error) {
	account, err := ts.accountRepository.Get(card)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.Account{}, ErrNoAccountFound
		} else {
			return domain.Account{}, err
		}
	}

	return account, nil
}

func (ts *TransactionService) CreateAccount(account domain.Account) (domain.Account, error) {
	return ts.accountRepository.Create(account)
}

func (ts *TransactionService) Create(transaction domain.Transaction) (domain.Transaction, error) {
	return ts.transactionRepository.Add(transaction)
}
