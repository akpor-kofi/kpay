package mailer

import (
	"fraud-detect-system/domain"
	"log"
	"net/smtp"
)

type MailtrapSMTP struct {
	Host     string
	Port     string
	Username string
	Password string
	Url      string
	Auth     smtp.Auth
}

func (s *MailtrapSMTP) Send(mail domain.Mail) error {
	from := "kofi.akpor@your.domain"

	err := smtp.SendMail(s.Url, s.Auth, from, []string{mail.To}, []byte(mail.Message))
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

func New(config *Config) *MailtrapSMTP {

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	url := config.Host + ":" + config.Port

	return &MailtrapSMTP{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		Url:      url,
		Auth:     auth,
	}
}
