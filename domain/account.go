package domain

import (
	"fraud-detect-system/domain/markov"
	"github.com/kamva/mgm/v3"
)

type Account struct {
	mgm.DefaultModel `bson:",inline"`
	CardNum          string `bson:"credit_card" json:"credit_card"` // TODO: make unique
	CardHolder       string `bson:"card_holder" json:"card_holder"`

	markov.HMM `bson:"markov_._hmm" json:"markov.HMM"`
}

// ClassifyTransaction TODO: REMOVE THIS
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
