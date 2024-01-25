package circuit

import (
	"S-gnark/frontend"
	"fmt"
	"math/big"
)

// mean = floor(sum / N), assume that mean = sum / N - e, 0 <= e < 1
// then N * mean = sum - N * e, which means sum - N * mean = N * e \in [0, N)
// get accuracy (N)
func computeMeanAccuracyHint(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	// no matter how many samples are given
	// here the accuracy is just related to N, so we just give input N
	if len(inputs) != 1 {
		return fmt.Errorf("ComputeMeanHint: input len must be 1")
	}
	// output is sum and mean, so len(outputs) = 1
	if len(outputs) != 1 {
		return fmt.Errorf("ComputeMeanHint: input len must be 1")
	}
	// just output <- N
	outputs[0].Add(outputs[0], inputs[0])
	return nil
}

// MeanCircuit
// verify the Mean which is claimed by prover is true in a high accuracy
type MeanCircuit struct {
	X [10000]frontend.Variable `gnark:"x"`
	N frontend.Variable        `gnark:",public"` // N is needed, if the number of sample is less than 10, we will fill 0
	Y frontend.Variable        `gnark:",public"`
}

func (circuit *MeanCircuit) Define(api frontend.API) error {
	/***
		Hints 这里需要考虑不能整除的情况
		并且考虑到给出的样本X,也就是X1, X2, ...., 可能本就不是整数
		因此，我们可以假设给定的X就是原有的100倍，来保证精度(小数点后两位）
		那么Y = 100E(X)
	***/
	// then \sigma X - Y * N = e * N, 0 <= e < 1, thus 0 <= e * N < N
	//solver.RegisterHint(computeMeanAccuracyHint)
	//accuracy, _ := api.Compiler().NewHint(computeMeanAccuracyHint, 1, circuit.N)
	// verify X - Y * N < N
	//api.AssertIsLessOrEqual(api.Sub(
	//	api.Add(circuit.X1, circuit.X2, circuit.X3, circuit.X4, circuit.X5, circuit.X6, circuit.X7, circuit.X8, circuit.X9, circuit.X10),
	//	api.Mul(circuit.Y, circuit.N)),
	//	circuit.N)
	api.AssertIsLessOrEqual(api.Sub(api.Add(circuit.X[0], circuit.X[1], circuit.X[2:]...), api.Mul(circuit.Y, circuit.N)), circuit.N)
	return nil
}
