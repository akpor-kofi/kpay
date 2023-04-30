package storage

import (
	"context"
	"errors"
	"fraud-detect-system/domain"
	"fraud-detect-system/markov"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNoDocument = errors.New("error: no document")
)

type AccountStorage struct {
	collName   string
	collection *mgm.Collection
	context    context.Context
}

func (a *AccountStorage) Get(creditCard string) (domain.Account, error) {
	var account domain.Account
	err := a.collection.FirstWithCtx(a.context, bson.M{"credit_card": creditCard}, &account)
	if err != nil {
		return domain.Account{}, err
	}

	return account, nil
}

func (a *AccountStorage) Create(account domain.Account) (domain.Account, error) {
	// initialize HMM
	account.HMM = *markov.New(markov.NumStates, markov.NumOfObservableSymbols, markov.MaxIters)

	err := a.collection.CreateWithCtx(a.context, &account)
	if err != nil {
		return domain.Account{}, err
	}

	return account, nil
}

func (a *AccountStorage) Save(account *domain.Account) error {
	return a.collection.UpdateWithCtx(a.context, account)
}

func NewAccountStorage(context context.Context, collName string) *AccountStorage {
	collection := mgm.CollectionByName(collName)

	_, err := collection.Indexes().CreateOne(context, mongo.IndexModel{
		Keys:    bson.D{{Key: "credit_card", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		panic(err)
	}

	return &AccountStorage{
		collName:   collName,
		collection: collection,
		context:    context,
	}
}
