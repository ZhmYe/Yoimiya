package circuit

import (
	"S-gnark/constraint/solver"
	"S-gnark/frontend"
	"fmt"
	"math/big"
)

// V = E(X^2) - (E(X))^2
// Y = V * n / (n - 1), this (n - 1)Y = nV
// so n * n * V = n * \sum(X^2) - sum(X)^2 = (n - 1)n * Y
// Y = floor((n * \sum(X^2) - sum(X) ^ 2) / (n - 1)n) = (n * \sum(X^2) - sum(X) ^ 2) / (n - 1)n - e, 0 <= e < 1
// so (n - 1)n * Y = n * \sum(X^2) - sum(X) ^ 2 - (n - 1)n * e
// get accuracy((n - 1) * n)
func computeSampleVarianceAccuracyHint(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	// no matter how many samples are given
	// here the accuracy is just related to N, so we just give input N
	if len(inputs) != 1 {
		return fmt.Errorf("ComputeMeanHint: input len must be 1")
	}
	// output is sum and mean, so len(outputs) = 1
	if len(outputs) != 1 {
		return fmt.Errorf("ComputeMeanHint: input len must be 1")
	}
	// just output <- N * (N -1)
	outputs[0].Mul(inputs[0], inputs[0].Sub(inputs[0], big.NewInt(1)))
	return nil
}

type SampleVarCircuit struct {
	// here we assume that sum of x(sample x_1,x_2, ..., x_n) is private input
	X1  frontend.Variable `gnark:"x1"`
	X2  frontend.Variable `gnark:"x2"`
	X3  frontend.Variable `gnark:"x3"`
	X4  frontend.Variable `gnark:"x4"`
	X5  frontend.Variable `gnark:"x5"`
	X6  frontend.Variable `gnark:"x6"`
	X7  frontend.Variable `gnark:"x7"`
	X8  frontend.Variable `gnark:"x8"`
	X9  frontend.Variable `gnark:"x9"`
	X10 frontend.Variable `gnark:"x10"`
	N   frontend.Variable `gnark:",public"`
	Y   frontend.Variable `gnark:",public"`
}

func (circuit *SampleVarCircuit) Define(api frontend.API) error {
	/***
		Hints: 样本方差 = \sigma (x - mean)^2 / (n - 1) = n * V / (n - 1)
	 ***/
	// we can verify n * V = (n - 1) * Y
	sum := api.Add(circuit.X1, circuit.X2, circuit.X3, circuit.X4, circuit.X5, circuit.X6, circuit.X7, circuit.X8, circuit.X9, circuit.X10)
	sumOfSquare := api.Add(api.Mul(circuit.X1, circuit.X1), api.Mul(circuit.X2, circuit.X2), api.Mul(circuit.X3, circuit.X3), api.Mul(circuit.X4, circuit.X4), api.Mul(circuit.X5, circuit.X5), api.Mul(circuit.X6, circuit.X6), api.Mul(circuit.X7, circuit.X7), api.Mul(circuit.X8, circuit.X8), api.Mul(circuit.X9, circuit.X9), api.Mul(circuit.X10, circuit.X10))
	// Y = (n * \sigma(X^2) - sigma(X)^2) / (n - 1)n, | dis | < (n - 1)n
	//dis := api.Sub(api.Sub(api.Mul(circuit.N, sumOfSquare), api.Mul(sum, sum)), api.Mul(api.Sub(circuit.N, 1), circuit.N, circuit.Y))
	solver.RegisterHint(computeSampleVarianceAccuracyHint)
	accuracy, _ := api.Compiler().NewHint(computeSampleVarianceAccuracyHint, 1, circuit.N)
	api.AssertIsLessOrEqual(
		api.Sub(api.Mul(circuit.N, sumOfSquare), api.Mul(sum, sum), api.Mul(circuit.N, api.Sub(circuit.N, 1), circuit.Y)),
		accuracy[0])
	return nil
}
