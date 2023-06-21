package fraud_detector_srv

import (
	"bytes"
	"encoding/json"
	"fraud-detect-system/domain"
	"fraud-detect-system/domain/markov"
	"fraud-detect-system/domain/ports"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
)

const (
	ErrCannotCluster = "no valid for clustering yet"
)

type FraudDetectorService struct {
	transactionRepository ports.TransactionRepository
	accountRepository     ports.AccountRepository
}

func New(accountRepository ports.AccountRepository, transactionRepository ports.TransactionRepository) *FraudDetectorService {
	return &FraudDetectorService{
		transactionRepository: transactionRepository,
		accountRepository:     accountRepository,
	}
}

func (fds *FraudDetectorService) cluster(points []float64, k int) []float64 {
	var d clusters.Observations

	for _, pt := range points {
		d = append(d, clusters.Coordinates{
			pt,
			1.0,
		})
	}

	km := kmeans.New()
	clusters, _ := km.Partition(d, k)

	var corePoints []float64

	for _, c := range clusters {
		corePoints = append(corePoints, c.Center[0])
	}

	//sort cluster points
	sort.Float64s(corePoints)

	return corePoints
}

func (fds *FraudDetectorService) classify(spendingHabits map[int]float64, amounts []float64) []int {
	var labeledData []int

	for _, amount := range amounts {
		switch {
		case amount >= spendingHabits[2]:
			labeledData = append(labeledData, 2)
		case amount >= spendingHabits[1]:
			labeledData = append(labeledData, 1)
		default:
			labeledData = append(labeledData, 0)
		}
	}

	return labeledData
}

func (fds *FraudDetectorService) DetectHMM(account *domain.Account) (HMMPred, error) {
	transactions, err := fds.transactionRepository.GetAllWhereAccountIs(account.CardNum)

	if err != nil {
		return HMMPred{}, err
	}

	// remove the current transaction
	transformedAmounts := transform(transactions)

	corePoints := fds.cluster(transformedAmounts, 3)

	clusteredAmounts := fds.classify(map[int]float64{
		0: corePoints[0], 1: corePoints[1], 2: corePoints[2]}, transformedAmounts)

	var obs [][]int

	currentTrans := clusteredAmounts[len(clusteredAmounts)-1]
	clusteredAmounts = clusteredAmounts[:len(clusteredAmounts)-2]

	window := 10

	for i := 0; i < len(clusteredAmounts)-window; i++ {
		obs = append(obs, clusteredAmounts[i:i+window])
	}

	hmm := fds.initializeAccountHMM()

	for i := 0; i < len(obs); i++ {
		hmm.Fit(obs[i])
	}

	prevSequence := obs[len(obs)-1]
	currSequence := append(prevSequence[1:], currentTrans)

	_, x, _ := hmm.DetectFraud(prevSequence)
	_, y, avgLikelihood := hmm.DetectFraud(currSequence)

	var isFraud bool

	if math.Abs(x-y) > 0.02 || math.IsNaN(x-y) {
		isFraud = true
	} else {
		isFraud = false
	}

	//if math.IsNaN(fraudProb) {
	//	fraudProb = 0
	//}

	return HMMPred{
		IsFraud:       isFraud,
		FraudProb:     y,
		AvgLikelihood: avgLikelihood,
	}, nil
}

func transform(txs []domain.Transaction) (amounts []float64) {
	for _, tx := range txs {
		amounts = append(amounts, tx.Amt)
	}
	return
}

func (fds *FraudDetectorService) initializeAccountHMM() markov.HMM {
	// initialize HMM
	return *markov.New(markov.NumStates, markov.NumOfObservableSymbols, markov.MaxIters)
}

func (fds *FraudDetectorService) DetectXGB(features XGBFeatures) (XGBPred, error) {
	return detectFraudXGB(features)
}

type EnsembleResult struct {
	IsHmmFraud bool    `json:"is_hmm_fraud"`
	Prob       float64 `json:"prob"`
	Likelihood float64 `json:"likelihood"`
}

func (fds *FraudDetectorService) Ensemble(hmmPred HMMPred, xgbPred XGBPred) EnsembleResult {
	ensemble := xgbPred.Probability[1]

	return EnsembleResult{
		IsHmmFraud: hmmPred.IsFraud,
		Prob:       ensemble,
		Likelihood: hmmPred.AvgLikelihood,
	}
}

type XGBFeatures struct {
	Hour                         int     `json:"hour"`
	Amount                       float64 `json:"amt"`
	CategoryIndex                int     `json:"categoryIndex"`
	TravelSpeed                  float64 `json:"travel_speed"`
	AvgSpendPw                   float64 `json:"avg_spend_pw"`
	LastDayTransactionCount      int     `json:"last_24h_transaction_count"`
	LastDayFraudTransactionCount int     `json:"last_24h_fraud_transaction_count"`
}

type XGBPred struct {
	Prediction  int       `json:"prediction"`
	Probability []float64 `json:"probability"`
}

type HMMPred struct {
	IsFraud       bool    `json:"is_fraud"`
	FraudProb     float64 `json:"fraud_prob"`
	AvgLikelihood float64 `json:"avg_likelihood"`
}

func detectFraudXGB(features XGBFeatures) (XGBPred, error) {
	url := os.Getenv("ML_SERVER_URL_DEV") + "predict/xgb"

	jsonData, err := json.Marshal(features)
	if err != nil {
		return XGBPred{}, err
	}

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return XGBPred{}, err
	}

	responseStream, err := io.ReadAll(resp.Body)
	if err != nil {
		return XGBPred{}, err
	}

	var xgbPred XGBPred
	print(string(responseStream))
	err = json.Unmarshal(responseStream, &xgbPred)
	if err != nil {
		return XGBPred{}, err
	}

	return xgbPred, nil
}

// should be able to do xgb

// should be able to do avg of xgb and hmm

// train hmm model -->

// create hmm -->

// cluster transactions -->

// reduce []Transactions --> []float64
