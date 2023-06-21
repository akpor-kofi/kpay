package clustering

import (
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/markov"
	"fraud-detect-system/domain/ports"
	"github.com/gofiber/fiber/v2"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"sort"
)

var (
	ErrCannotCluster = "no valid for clustering yet"
)

func ClusterTransactions(account domain.Account, transactionRepository ports.TransactionRepository, accountRepository ports.AccountRepository) error {
	// need to get amounts []float64

	transactions, err := transactionRepository.GetAllValidTransactionWhereAccountIs(account.CardNum)
	if err != nil {
		return err
	}

	if len(transactions) < 20 {
		return fiber.NewError(fiber.StatusOK, ErrCannotCluster)
	}

	var d clusters.Observations

	for _, amount := range transform(transactions[:len(transactions)-markov.NumOfObservations]) {
		d = append(d, clusters.Coordinates{
			amount,
			1.0, // till i find what else to weigh against amount
		})
	}

	km := kmeans.New()
	clusters, err := km.Partition(d, 3)
	if err != nil {
		return err
	}

	var corePoints []float64

	for _, c := range clusters {
		corePoints = append(corePoints, c.Center[0])
	}

	//sort cluster points
	sort.Float64s(corePoints)

	account.SpendingHabits[0] = corePoints[0] // TODO: make use of enum constant low, avg & high
	account.SpendingHabits[1] = corePoints[1]
	account.SpendingHabits[2] = corePoints[2]

	accountRepository.Save(&account)

	return nil
}

func transform(txs []domain.Transaction) (amounts []float64) {
	for _, tx := range txs {
		amounts = append(amounts, tx.Amt)
	}
	return
}
