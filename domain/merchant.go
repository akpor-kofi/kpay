package domain

import (
	"context"
	"errors"
	"github.com/kamva/mgm/v3"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Merchant struct {
	mgm.DefaultModel `bson:",inline"`
	Name             string    `bson:"name" json:"name" validate:"required"`
	Email            string    `bson:"email" json:"email" validate:"required,email"`
	Secret           string    `bson:"secret" json:"secret"` // TODO: don't add this
	Theme            string    `bson:"theme" json:"theme"`   // Not a string tho
	Password         string    `bson:"password" json:"password,omitempty" validate:"required"`
	ConfirmPassword  string    `bson:"confirm_password" json:"confirm_password,omitempty" validate:"required"`
	WebsiteUrl       string    `bson:"website_url" json:"website_url" validate:"required"`
	LastLoggedIn     time.Time `bson:"last_logged_in" json:"last_logged_in"`
	SiteInformation  string    `bson:"site_information" json:"site_information"`
}

func (m *Merchant) Creating(c context.Context) error {
	// hashed password
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(m.Password), 10)
	m.Password = string(hashedBytes)
	m.ConfirmPassword = ""

	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = time.Now().UTC()

	return nil
}

func (m *Merchant) Created(c context.Context) error {
	m.Password = ""
	return nil
}

func (m *Merchant) PasswordMatches(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password))
	if err != nil {
		return errors.New("username or password invalid")
	}

	return nil
}
