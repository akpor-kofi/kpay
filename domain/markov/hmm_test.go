package markov

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestHMM_Train(t *testing.T) {
	hmm := New(2, 3, 10)

	var observations [][]int
	//add other 20 more random observations

	// -------- SLIDING WINDOW -------
	var arr []int

	for j := 0; j < 10; j++ {
		rand.Seed(time.Now().UnixNano())
		//arr = append(arr, rand.Intn(2))
		arr = append(arr, 1)
	}

	prevArr := arr

	for i := 0; i < 20; i++ {
		slider := prevArr[1:]

		rand.Seed(time.Now().UnixNano())
		//slider = append(slider, rand.Intn(2))
		slider = append(slider, 1)

		prevArr = slider
		observations = append(observations, slider)
	}

	fmt.Println(observations)
	for i := 0; i < len(observations); i++ {
		hmm.Fit(observations[i])
		//hmm.Train(observations[i], 300)
	}

	//assert.Equal(t, math.Round(mat.Sum(&hmm.Pi)), 1.0, "should be equal")
	fmt.Println(hmm.A)

	fmt.Println("------------------------------------")
	fmt.Println(hmm.B)

	//fmt.Println(viterbi(hmm, []int{1, 2, 2, 1, 3, 1, 1, 1, 1, 1}))
	currentTrans := 2

	prevSequence := observations[len(observations)-1]
	currSequence := append(prevSequence[1:], currentTrans)

	_, x, _ := hmm.DetectFraud(prevSequence)
	_, y, _ := hmm.DetectFraud(currSequence)

	fmt.Println(x - y)
}
