package domain

import (
	"github.com/kamva/mgm/v3"
)

type Transaction struct {
	mgm.DefaultModel `bson:",inline"`
	Amt              float64 `bson:"amt" json:"amt" json:"amt"`
	Merchant         string  `bson:"merchant" json:"merchant" json:"merchant,omitempty"`
	UserAgent        string  `bson:"user_agent" json:"user_agent""`
	Ip               string  `bson:"ip_address" json:"ip_address" json:"ip,omitempty"`
	CreditCard       string  `bson:"credit_card" json:"credit_card" json:"credit_card,omitempty"`

	User
	//Product
	RiskScore float64 `bson:"risk_score" json:"risk_score"`
	IsFraud   bool    `bson:"is_fraud" json:"is_fraud,omitempty"`

	// ------ newly added ----------- (more data for the ml)
	StateOfTransaction string `bson:"state_of_transaction" json:"state_of_transaction"`
	SuspiciousIp       bool   `bson:"suspicious_ip" json:"suspicious_ip"` // basically if this ip has been used for a fraudulent transaction
	IsVpn              bool   `bson:"is_vpn" json:"is_vpn"`
	CardType           int    `bson:"card_type" json:"card_type"`

	// AGGREGATED
	// last_Transaction_24h_count | last_Transaction_24h_count | hour | avg_spend_pw | avg_spend_pm | travel_speed

	// BEHAVIOURAL
	NumberOfCardEntryAttempts int     `bson:"number_of_card_entry_attempts" json:"number_of_card_entry_attempts"`
	PaymentProcessDuration    float64 `bson:"payment_process_duration" json:"payment_process_duration"`
	NumberOfFailedAttempts    int     `bson:"number_of_failed_attempts" json:"number_of_failed_attempts"`

	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
}

type User struct {
	Email           string `bson:"email" json:"email" validate:"email,required"`
	JoinedAt        int    `bson:"joined_at" json:"joined_at"` // number of days
	EmailReputation bool   `bson:"email_reputation" json:"email_reputation"`

	// Optional
	ShoppingDuration      float64 `bson:"shopping_duration" json:"shopping_duration"`
	NumberOfSearchQueries float64 `bson:"number_of_search_queries" json:"number_of_search_queries"`

	// product quantity
	// product category
}
