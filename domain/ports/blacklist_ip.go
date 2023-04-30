package ports

import "fraud-detect-system/domain"

type BlacklistIpRepository interface {
	Add(blacklistedIp domain.BlacklistIp) error
	GetAll() ([]domain.BlacklistIp, error)
}
