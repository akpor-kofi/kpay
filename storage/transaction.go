package storage

import (
	"context"
	"fraud-detect-system/domain"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type TransactionStorage struct {
	collName   string
	collection *mgm.Collection
	context    context.Context
}

func (t *TransactionStorage) GetNewTransactions(merchantId string, lastLoggedIn time.Time) int {
	query := bson.M{
		"merchant":   merchantId,
		"created_at": bson.M{"$gte": lastLoggedIn},
	}

	totalNewNotifications, err := t.collection.CountDocuments(t.context, query)
	if err != nil {
		return 0
	}

	return int(totalNewNotifications)
}

func (t *TransactionStorage) GetRelatedPayments(merchantId string, prefix string) ([]domain.Transaction, error) {
	datePrefix := prefix + "T00:00:00Z"

	startDate, _ := time.Parse(time.RFC3339, datePrefix)

	hr, _, _ := startDate.Clock()

	remainingHour := time.Duration(24 - hr)
	duration := remainingHour * time.Hour

	endDate := startDate.Add(duration)

	query := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"_id": bson.M{operator.Regex: "^" + prefix}},
					{"ip_address": bson.M{operator.Regex: "^" + prefix}},
					{"state_of_transaction": bson.M{operator.Regex: "^" + prefix}},
					{"created_at": bson.M{operator.Gte: startDate, operator.Lte: endDate}},
				},
			},
			{"merchant": merchantId},
		},
	}

	var result []domain.Transaction

	err := t.collection.SimpleFind(&result, query)
	if err != nil {
		return []domain.Transaction{}, err
	}

	return result, nil
}

func (t *TransactionStorage) FetchTransactions(date time.Time, isFraud bool, merchant string) ([]domain.Transaction, error) {
	// Set the start and end time for the given date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Define the filter to query the transactions for the given date and isFraud flag
	filter := bson.M{
		"is_fraud": isFraud,
		"created_at": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
		"merchant": merchant,
	}

	// Execute the find query to get the transactions for the given date and isFraud flag
	cur, err := t.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	// Read the transactions
	transactions := make([]domain.Transaction, 0)
	for cur.Next(context.Background()) {
		var transaction domain.Transaction
		err := cur.Decode(&transaction)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (t *TransactionStorage) Add(newTransaction domain.Transaction) (domain.Transaction, error) {
	err := t.collection.CreateWithCtx(t.context, &newTransaction)
	if err != nil {
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

func (t *TransactionStorage) GetAllValidTransactionWhereAccountIs(creditCard string) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := t.collection.SimpleFindWithCtx(t.context, &transactions, bson.M{"credit_card": creditCard, "is_fraud": false})
	if err != nil {
		return []domain.Transaction{}, err
	}

	return transactions, nil
}

func (t *TransactionStorage) GetAllWhereMerchantIs(merchant string) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := t.collection.SimpleFindWithCtx(t.context, &transactions, bson.M{"merchant": merchant})
	if err != nil {
		return []domain.Transaction{}, err
	}

	return transactions, nil
}

func (t *TransactionStorage) GetAll() ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := t.collection.SimpleFindWithCtx(t.context, &transactions, bson.M{})
	if err != nil {
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
