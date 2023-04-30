package domain

import (
	"fraud-detect-system/markov"
	"github.com/kamva/mgm/v3"
)

type Account struct {
	mgm.DefaultModel `bson:",inline"`
	CardNum          string `bson:"credit_card" json:"credit_card"` // TODO: make unique
	CardHolder       string `bson:"card_holder" json:"card_holder"`
	IPAddress        string `bson:"ip_address" json:"ip_address"`
	Merchant         string `bson:"merchant" json:"merchant"`

	markov.HMM `bson:"markov_._hmm" json:"markov.HMM"`
}

func (a *Account) ClassifyTransaction(amount float64) int {
	switch {
	case amount >= a.SpendingHabits[2]:
		return 2
	case amount >= a.SpendingHabits[1]:
		return 1
	default:
		return 0
	}
}
