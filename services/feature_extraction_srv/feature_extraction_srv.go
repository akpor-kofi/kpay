package feature_extraction_srv

import (
	"errors"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/ports"
	"math"
	"time"
)

const (
	earthRadiusKm = 6371.0
)

type FeatureExtractionService struct {
	transactionRepository ports.TransactionRepository
}

func New(transactionRepository ports.TransactionRepository) *FeatureExtractionService {
	return &FeatureExtractionService{
		transactionRepository: transactionRepository,
	}
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	lat1 = degreesToRadians(lat1)
	lng1 = degreesToRadians(lng1)
	lat2 = degreesToRadians(lat2)
	lng2 = degreesToRadians(lng2)

	deltaLat := lat2 - lat1
	deltaLon := lng2 - lng1

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadiusKm * c

	return distance
}

func (fes *FeatureExtractionService) ExtractLast24HourCount(currTransaction domain.Transaction) (int, int) {
	// Assuming transactions is a slice of Transaction structs and currTransaction is the current transaction
	transactions, _ := fes.getAllPreviousTransactions(currTransaction)

	last24hTransactionCount := 0
	last24hFraudTransactionCount := 0

	for _, transaction := range transactions {
		// Assuming you have a field representing the transaction date (e.g., transDate)
		if currTransaction.CreatedAt.Sub(transaction.CreatedAt) <= 24*time.Hour {
			last24hTransactionCount++

			// Assuming you have a field indicating fraud (e.g., IsFraud)
			if transaction.IsFraud {
				last24hFraudTransactionCount++
			}
		}
	}

	return last24hTransactionCount, last24hFraudTransactionCount
}

func (fes *FeatureExtractionService) getAllPreviousTransactions(transaction domain.Transaction) ([]domain.Transaction, error) {
	cardNumber := transaction.CreditCard

	return fes.transactionRepository.GetAllWhereAccountIs(cardNumber)
}

func (fes *FeatureExtractionService) getPreviousTransaction(transaction domain.Transaction) (domain.Transaction, error) {

	cardNumber := transaction.CreditCard

	transactions, err := fes.transactionRepository.GetAllWhereAccountIs(cardNumber)

	if err != nil {
		return domain.Transaction{}, err
	}

	if len(transactions) <= 1 {
		return domain.Transaction{}, errors.New("not enough transactions")
	}

	return transactions[len(transactions)-2], nil
}

func (fes *FeatureExtractionService) ExtractHour(currTransaction domain.Transaction) int {
	return currTransaction.CreatedAt.Hour()
}

func (fes *FeatureExtractionService) ExtractAvgSpend(currTransaction domain.Transaction) float64 {
	// Assuming you have access to the list of transactions for a credit card
	transactions, _ := fes.getAllPreviousTransactions(currTransaction)

	totalAmount := 0.0
	numWeeks := 0

	// Iterate over transactions and calculate the total amount and number of weeks
	for _, transaction := range transactions {
		totalAmount += transaction.Amt
		// Assuming you have a field representing the transaction date (e.g., transDate)
		if transaction.CreatedAt.Weekday() == time.Sunday {
			numWeeks++
		}
	}

	if numWeeks == 0 {
		numWeeks++
	}

	avgSpendPw := totalAmount / float64(numWeeks)
	return avgSpendPw
}

func (fes *FeatureExtractionService) ExtractTravelSpeed(currentTransaction domain.Transaction) float64 {
	// Assuming prevTransaction and currTransaction are instances of Transaction

	//TODO: remove this
	if currentTransaction.Ip == "127.0.0.1" {
		return 0.5
	}
	previousTransaction, err := fes.getPreviousTransaction(currentTransaction)

	if errors.Is(err, errors.New("not enough transactions")) {
		return 0
	}

	distance := calculateDistance(previousTransaction.Latitude, previousTransaction.Longitude, currentTransaction.Latitude, currentTransaction.Longitude)
	timeDifference := currentTransaction.CreatedAt.Sub(previousTransaction.CreatedAt)

	travelSpeed := distance / timeDifference.Hours()

	return travelSpeed
}
