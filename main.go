package main

import (
	"encoding/csv"
	"fmt"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/evaluation"
	"github.com/sjwhitworth/golearn/knn"
	"os"
	"reflect"
	"unsafe"
)

func mainx() {
	// ------- Monitor a user transaction -----

	//user := new(domain.User)
	//err := mgm.Coll(user).FindByID("63b399e1739d74c2dc404ea2", user)
	//check(err)
	//
	//fmt.Println(user.Balance)
	//
	//trans := new(domain.Transaction)
	//var results []domain.Transaction
	//
	//opts := options.Find().SetSort(bson.D{{"amt", 1}})
	//check(mgm.Coll(trans).SimpleFind(&results, bson.M{"user._id": user.ID}, opts))
	//
	//fmt.Println(len(results))
	//
	//user.HMM = domain.NewHMM()
	//
	//user.AddObservations(results[12:22], results[:12])
	//
	//user.Train(20)
	//
	//err = mgm.Coll(user).Update(user)
	//check(err)

	//inst, err := base.ParseCSVToInstances("shorted_data.csv", true)
	//if err != nil {
	//	return
	//}
	//
	//train, test := base.InstancesTrainTestSplit(inst, 0.50)
	//
	//rand.Seed(time.Now().Unix())
	//
	//iforest := trees.NewIsolationForest(200, 100, 200)
	//iforest.Fit(train)
	//
	//preds := iforest.Predict(test)
	//
	//for i, pred := range preds {
	//	if pred > 0.7 {
	//		fmt.Println("outlier on row ", i, " with anomaly score of ", pred)
	//	}
	//}

	//shortenBigCSVFile("fraudTrain.csv", check, 70000, 2, 3, 4, 5, 13, 14, 22)

	//incomingData := base.NewDenseCopy(train)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func oneHotEncode(data, categories []string) [][]int {
	oneHotEncoding := make(map[string][]int)

	// Initialize the one hot encoded values for each category
	for i, category := range categories {
		oneHotEncoding[category] = make([]int, len(categories))
		oneHotEncoding[category][i] = 1
	}

	// One hot encode the data
	encodedData := make([][]int, len(data))
	for i, datum := range data {
		encodedData[i] = oneHotEncoding[datum]
	}

	return encodedData
}

func oneHotDecode(encodedData [][]int, categories []string) []string {
	// Create a map to store the inverse one hot encoding for each category
	inverseOneHotEncoding := make(map[string][]int)

	// Initialize the inverse one hot encoding for each category
	for i, category := range categories {
		encoding := make([]int, len(categories))
		encoding[i] = 1
		inverseOneHotEncoding[category] = encoding
	}

	// Decode the one hot encoded data
	decodedData := make([]string, len(encodedData))
	for i, encodedDatum := range encodedData {
		for category, encoding := range inverseOneHotEncoding {
			if reflect.DeepEqual(encodedDatum, encoding) {
				decodedData[i] = category
				break
			}
		}
	}

	return decodedData
}

func shortenBigCSVFile(name string, errorChecker func(error), limit int, cols ...int) {
	file, err := os.Open(name)
	errorChecker(err)

	r := csv.NewReader(file)

	records, err := r.ReadAll()

	fmt.Println("done")

	fmt.Println(records[0])

	var newR [][]string

	for i := 0; i < limit; i++ {
		var record []string

		for _, j := range cols {
			record = append(record, records[i][j])
		}
		newR = append(newR, record)
	}

	newFile, err := os.Create("shorted_" + name)
	if err != nil {
		return
	}

	w := csv.NewWriter(newFile)

	err = w.WriteAll(newR)
	errorChecker(err)
}

func unpackBytesToFloat(val []byte) float64 {
	pb := unsafe.Pointer(&val[0])
	return *(*float64)(pb)
}

func knnCls(test base.FixedDataGrid) {
	cls := knn.NewKnnClassifier("euclidean", "linear", 8)

	err := cls.Load("kofi.pkl")
	if err != nil {
		panic(err)
	}

	predictions, err := cls.Predict(test)
	if err != nil {
		panic(err)
	}

	//fmt.Println(predictions)

	confusionMat, err := evaluation.GetConfusionMatrix(test, predictions)
	if err != nil {
		panic(fmt.Sprintf("Unable to get confusion matrix: %s", err.Error()))
	}
	fmt.Println(evaluation.GetSummary(confusionMat))
}
