package storage

import (
	"context"
	"fmt"
	"fraud-detect-system/domain"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlacklistStorage struct {
	collName   string
	collection *mgm.Collection
	context    context.Context
}

func NewBlacklistStorage(context context.Context, collName string) *BlacklistStorage {
	collection := mgm.CollectionByName(collName)

	_, err := collection.Indexes().CreateOne(context, mongo.IndexModel{
		Keys:    bson.D{{Key: "ip_address", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		panic(err)
	}

	return &BlacklistStorage{
		collName:   collName,
		collection: collection,
		context:    context,
	}
}

func (bs *BlacklistStorage) Add(ip domain.BlacklistIp) error {
	err := bs.collection.CreateWithCtx(bs.context, &ip)
	if err != nil {
		fmt.Println("line 19 error storage.go")
		return err
	}

	return nil
}

func (bs *BlacklistStorage) GetAll() ([]domain.BlacklistIp, error) {
	var blacklistedIps []domain.BlacklistIp
	err := bs.collection.SimpleFindWithCtx(bs.context, &blacklistedIps, bson.M{})
	if err != nil {
		fmt.Println("line 41 storage.go")
		return []domain.BlacklistIp{}, err
	}

	return blacklistedIps, nil
}
