package ports

import (
	"fraud-detect-system/domain"
)

type MerchantRepository interface {
	Add(newMerchant *domain.Merchant) (domain.Merchant, error)
	GetById(id string) (domain.Merchant, error)
	GetByIdAndSecret(id, secret string) (domain.Merchant, error)
	GetByEmail(email string) (domain.Merchant, error)
	Save(merchant *domain.Merchant) error
}
