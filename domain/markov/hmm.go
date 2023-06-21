package markov

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const fraudThreshold = 0.4

const (
	NumStates              = 2
	NumOfObservableSymbols = 3
	//NumOfObservations = 10
	NumOfObservations = 10
	MaxIters          = 100
)

const (
	notFraud = iota
	fraud
)

const (
	low = iota
	average
	high
)

type HMM struct {
	N              int             `bson:"n" json:"n,omitempty"`
	M              int             `bson:"m" json:"m,omitempty"`
	MaxIters       int             `bson:"max_iters" json:"max_iters,omitempty"`
	LogProb        float64         `bson:"log_prob" json:"log_prob"`
	OldLogProb     float64         `bson:"old_log_prob" json:"old_log_prob"`
	Pi             []float64       `bson:"pi" json:"pi"`
	A              [][]float64     `bson:"a" json:"a"`
	B              [][]float64     `bson:"b" json:"b"`
	SpendingHabits map[int]float64 `bson:"spending_habits" json:"spending_habits,omitempty"`
}

func New(N, M, MaxIters int) *HMM {
	rand.Seed(time.Now().UnixNano())
	Pi := make([]float64, N)
	sum := 0.0
	for i := range Pi {
		Pi[i] = rand.Float64()
		sum += Pi[i]
	}

	for i := range Pi {
		Pi[i] /= sum
	}

	A := make([][]float64, N)

	for i := range A {
		A[i] = make([]float64, N)
		sumA := 0.0
		for j := range A[i] {
			A[i][j] = rand.Float64()
			sumA += A[i][j]
		}

		for j := range A[i] {
			A[i][j] /= sumA
		}
	}

	B := make([][]float64, N)

	for i := range B {
		B[i] = make([]float64, M)
		sumB := 0.0
		for j := range B[i] {
			B[i][j] = rand.Float64()
			sumB += B[i][j]
		}

		for j := range B[i] {
			B[i][j] /= sumB
		}
	}

	return &HMM{
		N:              N,
		M:              M,
		MaxIters:       MaxIters,
		Pi:             Pi,
		A:              A,
		B:              B,
		OldLogProb:     math.Inf(-1),
		SpendingHabits: make(map[int]float64),
	}
}

func (h *HMM) Fit(obs []int) {
	var iter int

	alpha, c := h.forwardPass(obs)
	beta := h.backwardPass(obs, c)

	gamma, psi := h.computeGamma(obs, alpha, beta)

	h.OldLogProb = math.Inf(-1)
	for h.next(&iter) {
		fmt.Println(iter)
		h.estimate(obs, gamma, psi)
	}

}

func (h *HMM) forwardPass(obs []int) ([][]float64, []float64) {
	T := len(obs)

	alpha := make([][]float64, T)
	c := make([]float64, T)

	c[0] = 0
	for i := range alpha {
		alpha[i] = make([]float64, h.N)
	}

	// initialize alpha
	for i := 0; i < h.N; i++ {
		alpha[0][i] = h.Pi[i] * h.B[i][obs[0]]
		c[0] += alpha[0][i]
	}

	// scale alpha[0]
	c[0] = 1.0 / c[0]
	for i := 0; i < h.N; i++ {
		alpha[0][i] *= c[0]
	}

	for t := 1; t < T; t++ {
		c[t] = 0
		for i := 0; i < h.N; i++ {
			alpha[t][i] = 0
			for j := 0; j < h.N; j++ {
				alpha[t][i] += alpha[t-1][j] * h.A[i][j]
			}
			alpha[t][i] *= h.B[i][obs[t]]
			c[t] += alpha[t][i]
		}

		// scale alpha[t]
		c[t] = 1.0 / c[t]
		for i := 0; i < h.N; i++ {
			alpha[t][i] *= c[t]
		}
	}

	h.computeLogProb(c, T)

	return alpha, c
}

func (h *HMM) backwardPass(obs []int, c []float64) [][]float64 {
	T := len(obs)

	beta := make([][]float64, T)

	for i := range beta {
		beta[i] = make([]float64, h.N)
	}

	// initialize beta
	for i := 0; i < h.N; i++ {
		beta[T-1][i] = c[T-1]
	}

	for t := T - 2; t > -1; t-- {
		for i := 0; i < h.N; i++ {
			beta[t][i] = 0
			for j := 0; j < h.N; j++ {
				beta[t][i] += h.A[i][j] * h.B[j][obs[t+1]] * beta[t+1][j]
			}

			// scale beta
			beta[t][i] *= c[t]
		}
	}

	return beta
}

func (h *HMM) computeGamma(obs []int, alpha, beta [][]float64) ([][][]float64, [][]float64) {
	T := len(obs)

	gamma := make([][][]float64, T)

	for i := range gamma {
		gamma[i] = make([][]float64, h.N)
		for j := range gamma[i] {
			gamma[i][j] = make([]float64, h.N)
		}
	}

	psi := make([][]float64, T)

	for i := range psi {
		psi[i] = make([]float64, h.N)
	}

	for t := 0; t < T-2; t++ {
		for i := 0; i < h.N; i++ {
			psi[t][i] = 0
			for j := 0; j < h.N; j++ {
				gamma[t][i][j] = alpha[t][i] * h.A[i][j] * h.B[j][obs[t+1]] * beta[t+1][j]
				psi[t][i] += gamma[t][i][j]
			}
		}
	}

	for i := 0; i < h.N; i++ {
		psi[T-1][i] = alpha[T-1][i]
	}

	return gamma, psi
}

func (h *HMM) estimate(obs []int, gamma [][][]float64, psi [][]float64) ([]float64, [][]float64, [][]float64) {
	T := len(obs)
	// estimate pi
	for i := 0; i < h.N; i++ {
		h.Pi[i] = psi[0][i]
	}

	// estimate A
	for i := 0; i < h.N; i++ {
		denom := 0.0
		for t := 0; t < T-1; t++ { // ****
			denom += psi[t][i]
		}

		for j := 0; j < h.N; j++ {
			numer := 0.0
			for t := 0; t < T-1; t++ {
				numer += gamma[t][i][j]
			}
			h.A[i][j] = numer / denom
		}
	}

	// estimate B
	for i := 0; i < h.N; i++ {
		denom := 0.0
		for t := 0; t < T; t++ { //****
			denom += psi[t][i]
		}

		for j := 0; j < h.M; j++ {
			numer := 0.0
			for t := 0; t < T; t++ {
				if obs[t] == j {
					numer += psi[t][i]
				}
			}

			h.B[i][j] = numer / denom
		}
	}

	return h.Pi, h.A, h.B
}

func (h *HMM) computeLogProb(c []float64, T int) float64 {
	logProb := 0.0

	for i := 0; i < T; i++ {
		logProb += math.Log(c[i])
	}

	logProb = -logProb

	h.LogProb = logProb

	return logProb
}

func (h *HMM) next(iter *int) bool {
	*iter++

	if *iter < h.MaxIters && h.LogProb > h.OldLogProb {
		h.OldLogProb = h.LogProb
		return true
	} else {
		return false
	}
}

func (h *HMM) DetectFraud(obs []int) (isFraud bool, maxProb float64, avgLikelihood float64) {
	viterbiSeq, _ := h.viterbi(obs)

	fmt.Println("obs: ", obs)
	fmt.Println("viberti sequence: ", viterbiSeq)

	T := len(viterbiSeq)

	// calculate the percentage of the seq that are 1, if more than 50% (0.50) likely fraud

	numOfFraudState := 0
	for _, o := range viterbiSeq {
		if o == fraud {
			numOfFraudState++
		}
	}

	p := float64(numOfFraudState) / float64(T)

	if viterbiSeq[T-1] == 1 {
		isFraud = true
	} else {
		isFraud = false
	}

	maxProb = h.LogLikelihood(viterbiSeq)

	avgLikelihood = p

	fmt.Println("*** ", h.LogLikelihood(viterbiSeq))

	return
}

func (h *HMM) viterbi(obs []int) ([]int, float64) {
	T := len(obs)
	delta := make([][]float64, T)
	psi := make([][]int, T)

	// initialization
	delta[0] = make([]float64, h.N)
	for i := 0; i < h.N; i++ {
		delta[0][i] = math.Log(h.Pi[i]) + math.Log(h.B[i][obs[0]])
		psi[0] = make([]int, h.N)
	}

	// recursion
	for t := 1; t < T; t++ {
		delta[t] = make([]float64, h.N)
		psi[t] = make([]int, h.N)
		for j := 0; j < h.N; j++ {
			maxDelta := -math.MaxFloat64
			maxIndex := 0
			for i := 0; i < h.N; i++ {
				deltaIJ := delta[t-1][i] + math.Log(h.A[i][j])
				if deltaIJ > maxDelta {
					maxDelta = deltaIJ
					maxIndex = i
				}
			}
			delta[t][j] = maxDelta + math.Log(h.B[j][obs[t]])
			psi[t][j] = maxIndex
		}
	}

	// termination
	maxProb := -math.MaxFloat64
	maxIndex := 0
	for i := 0; i < h.N; i++ {
		if delta[T-1][i] > maxProb {
			maxProb = delta[T-1][i]
			maxIndex = i
		}
	}

	// backtracking
	stateSeq := make([]int, T)
	stateSeq[T-1] = maxIndex
	for t := T - 2; t >= 0; t-- {
		stateSeq[t] = psi[t+1][stateSeq[t+1]]
	}

	return stateSeq, math.Exp(maxProb)

}

func (h *HMM) LogLikelihood(x []int) float64 {
	// Returns log P(x | model)
	// using the forward part of the forward-backward algorithm
	T := len(x)
	scale := make([]float64, T)
	alpha := make([][]float64, T)
	for i := range alpha {
		alpha[i] = make([]float64, h.M)
	}
	alpha[0] = elementWiseMultiply(h.Pi, getColumn(h.B, x[0]))
	scale[0] = sum(alpha[0])
	alpha[0] = elementWiseDivide(alpha[0], scale[0])
	for t := 1; t < T; t++ {
		alphaTPrime := dotProduct(alpha[t-1], h.A)
		for i := range alphaTPrime {
			alphaTPrime[i] *= h.B[i][x[t]]
		}
		scale[t] = sum(alphaTPrime)
		alpha[t] = elementWiseDivide(alphaTPrime, scale[t])
	}
	return math.Log(sum(scale))
}

// Helper functions

func elementWiseMultiply(a, b []float64) []float64 {
	result := make([]float64, len(a))
	for i := range a {
		result[i] = a[i] * b[i]
	}
	return result
}

func elementWiseDivide(a []float64, b float64) []float64 {
	result := make([]float64, len(a))
	for i := range a {
		result[i] = a[i] / b
	}
	return result
}

func dotProduct(a []float64, b [][]float64) []float64 {
	result := make([]float64, len(b[0]))
	for i := range b[0] {
		sum := 0.0
		for j := range a {
			sum += a[j] * b[j][i]
		}
		result[i] = sum
	}
	return result
}

func sum(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum
}

func getColumn(matrix [][]float64, columnIndex int) []float64 {
	column := make([]float64, len(matrix))
	for i := range matrix {
		column[i] = matrix[i][columnIndex]
	}
	return column
}
