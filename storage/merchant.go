package storage

import (
	"context"
	"fmt"
	"fraud-detect-system/domain"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MerchantStorage struct {
	collName   string
	collection *mgm.Collection
	context    context.Context
}

func (m MerchantStorage) Add(newMerchant domain.Merchant) (domain.Merchant, error) {
	err := m.collection.CreateWithCtx(m.context, &newMerchant)
	if err != nil {
		fmt.Println("line 19 error storage.go")
		return domain.Merchant{}, err
	}

	return newMerchant, err
}

func (m MerchantStorage) GetById(id string) (domain.Merchant, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.Merchant{}, err
	}

	var result domain.Merchant

	err = m.collection.FindByID(objectId, &result)
	if err != nil {
		return domain.Merchant{}, err
	}

	return result, err
}

func (m MerchantStorage) GetByIdAndSecret(id, secret string) (domain.Merchant, error) {
	// TODO: IMPLEMENT THIS

	return domain.Merchant{}, nil
}

func NewMerchantStorage(ctx context.Context, collName string) *MerchantStorage {
	collection := mgm.CollectionByName(collName)

	return &MerchantStorage{
		collName:   collName,
		collection: collection,
		context:    ctx,
	}
}
