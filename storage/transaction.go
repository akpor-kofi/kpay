package storage

import (
	"context"
	"fmt"
	"fraud-detect-system/domain"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
)

type TransactionStorage struct {
	collName   string
	collection *mgm.Collection
	context    context.Context
}

func (t *TransactionStorage) Add(newTransaction domain.Transaction) (domain.Transaction, error) {
	err := t.collection.CreateWithCtx(t.context, &newTransaction)
	if err != nil {
		fmt.Println("line 19 error storage.go")
		return domain.Transaction{}, err
	}

	return newTransaction, err
}

func (t *TransactionStorage) GetAllWhereAccountIs(creditCard string) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := t.collection.SimpleFindWithCtx(t.context, &transactions, bson.M{"credit_card": creditCard})
	if err != nil {
		return []domain.Transaction{}, err
	}

	return transactions, nil
}

func (t *TransactionStorage) GetAll() ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := t.collection.SimpleFindWithCtx(t.context, &transactions, bson.M{})
	if err != nil {
		fmt.Println("line 41 storage.go")
		return []domain.Transaction{}, err
	}

	return transactions, nil
}

func (t *TransactionStorage) Save(transaction *domain.Transaction) error {
	return t.collection.UpdateWithCtx(t.context, transaction)
}

func NewTransactionStorage(ctx context.Context, collName string) *TransactionStorage {
	collection := mgm.CollectionByName(collName)

	return &TransactionStorage{
		collName:   collName,
		collection: collection,
		context:    ctx,
	}
}
