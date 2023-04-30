package domain

import (
	"github.com/kamva/mgm/v3"
)

type Transaction struct {
	mgm.DefaultModel `bson:",inline"`
	Amt              float64 `bson:"amt" json:"amt" json:"amt,omitempty"`
	Merchant         string  `bson:"merchant" json:"merchant" json:"merchant,omitempty"`
	UserAgent        string  `bson:"user_agent" json:"user_gent""`
	Ip               string  `bson:"ip_address" json:"ip_address" json:"ip,omitempty"`
	CreditCard       string  `bson:"credit_card" json:"credit_card" json:"credit_card,omitempty"`
	User
	//Product
	RiskScore float64 `bson:"risk_score" json:"risk_score"`
	IsFraud   bool    `bson:"is_fraud" json:"is_fraud,omitempty"`
}

// type Product struct {
//	Category string `bson:"category" json:"category"`
//}

type User struct {
	Email string `bson:"email" json:"email" validate:"email,required"`
	Name  string `bson:"name" json:"name" validate:"required"`
	Phone string `bson:"phone" json:"phone"`
}
