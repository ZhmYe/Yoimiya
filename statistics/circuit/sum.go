package circuit

import "Yoimiya/frontend"

// SumCircuit
// circuit
// use slice to obtain n elements ?
// todo 从实践来看，似乎好像circuit不支持slice，也就是可变数量的x[]
// here we set the number of sample in the circuit 10
// if the number of sample is greater than 10, then we use several circuits compute in parallel, and then check all the proofs in a circuit.
// todo 验证多个proof的电路还需要写，暂时简写为多个verifier验证为true
type SumCircuit struct {
	// here we assume that x(sample x_1,x_2, ..., x_n) is private input
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
	//N frontend.Variable `gnark:",public"`
	Y frontend.Variable `gnark:",public"`
}

func (circuit *SumCircuit) Define(api frontend.API) error {
	api.AssertIsEqual(circuit.Y, api.Add(circuit.X1, circuit.X2, circuit.X3, circuit.X4, circuit.X5, circuit.X6, circuit.X7, circuit.X8, circuit.X9, circuit.X10))
	return nil
}
