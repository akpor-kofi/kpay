package markov

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestHMM_Train(t *testing.T) {
	hmm := New(2, 3, 100)

	var observations [][]int
	//add other 20 more random observations
	for i := 0; i < 30; i++ {
		var arr []int

		for j := 0; j < 10; j++ {
			rand.Seed(time.Now().UnixNano())
			arr = append(arr, rand.Intn(2))
			//arr = append(arr, 1)
		}

		observations = append(observations, arr)
	}

	observations = append(observations, []int{0, 0, 1, 1, 1, 1, 1, 1, 1, 1})
	for i := 0; i < len(observations); i++ {
		hmm.Fit(observations[i])
		//hmm.Train(observations[i], 300)
	}

	//assert.Equal(t, math.Round(mat.Sum(&hmm.Pi)), 1.0, "should be equal")
	fmt.Println(hmm.A)

	fmt.Println("--------------------------------")
	fmt.Println(hmm.B)

	//fmt.Println(viterbi(hmm, []int{1, 2, 2, 1, 3, 1, 1, 1, 1, 1}))
	fmt.Println(hmm.DetectFraud([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}))
}
