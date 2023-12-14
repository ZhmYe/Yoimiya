package circuit

import (
	"S-gnark/constraint/solver"
	"S-gnark/frontend"
	"fmt"
	"math/big"
)

// variance = floor(E(X^2)) - floor(E(X)^2), assume that variance = E(X^2) - E(X)^2 - e, 0 <= e < 1
// then N^2 * variance = N * sum(X^2) - sum(X)^2 - N^2 * e, which means
// N * sum(X^2) - sum(X)^2 - N^2 * variance = N^2 * e \in [0, N^2)
// get accuracy (N^2)
func computeVarianceAccuracyHint(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	// no matter how many samples are given
	// here the accuracy is just related to N, so we just give input N
	if len(inputs) != 1 {
		return fmt.Errorf("ComputeMeanHint: input len must be 1")
	}
	// output is sum and mean, so len(outputs) = 1
	if len(outputs) != 1 {
		return fmt.Errorf("ComputeMeanHint: input len must be 1")
	}
	// just output <- N^2
	outputs[0].Mul(inputs[0], inputs[0])
	return nil
}

type VarCircuit struct {
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

func (circuit *VarCircuit) Define(api frontend.API) error {
	/***
		Hints: 计算方差的第一种方式
		V(X) = \sigma (x_i - E(X))^2
		这种方式误差不好估计
	***/
	// first, we can design a circuit to compute the sum of the new sample above
	// compute \sigma(X)
	sum := api.Add(circuit.X1, circuit.X2, circuit.X3, circuit.X4, circuit.X5, circuit.X6, circuit.X7, circuit.X8, circuit.X9, circuit.X10)
	/***
		Hints: 计算方差的第二种方式
		V(X) = E(X^2) - (E(X))^2
	 ***/
	// then we compute \sigma(X^2)
	sumOfSquare := api.Add(api.Mul(circuit.X1, circuit.X1), api.Mul(circuit.X2, circuit.X2), api.Mul(circuit.X3, circuit.X3), api.Mul(circuit.X4, circuit.X4), api.Mul(circuit.X5, circuit.X5), api.Mul(circuit.X6, circuit.X6), api.Mul(circuit.X7, circuit.X7), api.Mul(circuit.X8, circuit.X8), api.Mul(circuit.X9, circuit.X9), api.Mul(circuit.X10, circuit.X10))
	// to have a higher accuracy
	// V(X) = sumOfSquare / N + (sum / N) ^ 2
	// we can verify by
	// N^2V(X) = N * sumOfSquare + sum^2
	/***
		Hints: 外面传进来的V(X)是取整后的结果
		如果V(X) = floor(E(X^2)) - (floor(E(X)))^2，也就是后面一部分先取整再平方，误差会很大
		这里是V(X) = floor(E(X^2)) - floor(E(X)^2),这样误差可以控制在(-n^2, n^2), |dis| < n^2来验证
	***/
	solver.RegisterHint(computeVarianceAccuracyHint)
	accuracy, _ := api.Compiler().NewHint(computeVarianceAccuracyHint, 1, circuit.N)
	api.AssertIsLessOrEqual(api.Sub(api.Mul(sumOfSquare, circuit.N), api.Mul(sum, sum), api.Mul(circuit.N, circuit.N, circuit.Y)),
		accuracy[0])
	return nil
}
