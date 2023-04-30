package ports

import "fraud-detect-system/domain"

type AccountRepository interface {
	Get(creditCard string) (domain.Account, error)
	Create(account domain.Account) (domain.Account, error)
	Save(account *domain.Account) error
}
