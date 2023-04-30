package domain

import "github.com/kamva/mgm/v3"

type Merchant struct {
	mgm.DefaultModel  `bson:",inline"`
	Name              string `bson:"name" json:"name" validate:"required"`
	Email             string `bson:"email" json:"email" validate:"required,email"`
	ProfilePictureUrl string `bson:"profile_picture_url" json:"profile_picture_url"` // give a default image
	Secret            string `bson:"secret" json:"secret"`                           // TODO: don't add this
	Theme             string `bson:"theme" json:"theme"`                             // Not a string tho
	Password          string `bson:"password" json:"password" validate:"required"`
	ConfirmPassword   string `bson:"confirm_password" json:"confirm_password" validate:"required"`
	WebsiteUrl        string `bson:"website_url" json:"website_url" validate:"required"`
	CallbackUrl       string `bson:"callback_url" json:"callback_url"`
	SiteInformation   string `bson:"site_information" json:"site_information"`
}
