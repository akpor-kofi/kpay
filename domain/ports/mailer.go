package ports

import "fraud-detect-system/domain"

type IMailer interface {
	Send(mail domain.Mail) error
}
