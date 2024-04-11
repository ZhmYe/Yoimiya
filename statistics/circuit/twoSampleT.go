package circuit

import (
	"Yoimiya/frontend"
)

type TwoSampleTCircuit struct {
	// here we assume that x(sample x_1,x_2, ..., x_n) is private input, y is public
	//X [10]int `gnark:"x"`
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
	//Y [10]int `gnark:",public"`
	Y1  frontend.Variable `gnark:",public"`
	Y2  frontend.Variable `gnark:",public"`
	Y3  frontend.Variable `gnark:",public"`
	Y4  frontend.Variable `gnark:",public"`
	Y5  frontend.Variable `gnark:",public"`
	Y6  frontend.Variable `gnark:",public"`
	Y7  frontend.Variable `gnark:",public"`
	Y8  frontend.Variable `gnark:",public"`
	Y9  frontend.Variable `gnark:",public"`
	Y10 frontend.Variable `gnark:",public"`
	M1  frontend.Variable `gnark:",public"`
	M2  frontend.Variable `gnark:",public"`
	T   frontend.Variable `gnark:", public"`
	//N2 frontend.Variable `gnark:",public"`
}

func (circuit *TwoSampleTCircuit) Define(api frontend.API) error {
	sumX := api.Add(circuit.X1, circuit.X2, circuit.X3, circuit.X4, circuit.X5, circuit.X6, circuit.X7, circuit.X8, circuit.X9, circuit.X10)
	sumY := api.Add(circuit.Y1, circuit.Y2, circuit.Y3, circuit.Y4, circuit.Y5, circuit.Y6, circuit.Y7, circuit.Y8, circuit.Y9, circuit.Y10)
	SquareSumX := api.Add(api.Mul(circuit.X1, circuit.X1), api.Mul(circuit.X2, circuit.X2), api.Mul(circuit.X3, circuit.X3), api.Mul(circuit.X4, circuit.X4), api.Mul(circuit.X5, circuit.X5), api.Mul(circuit.X6, circuit.X6), api.Mul(circuit.X7, circuit.X7), api.Mul(circuit.X8, circuit.X8), api.Mul(circuit.X9, circuit.X9), api.Mul(circuit.X10, circuit.X10))
	SquareSumY := api.Add(api.Mul(circuit.Y1, circuit.Y1), api.Mul(circuit.Y2, circuit.Y2), api.Mul(circuit.Y3, circuit.Y3), api.Mul(circuit.Y4, circuit.Y4), api.Mul(circuit.Y5, circuit.Y5), api.Mul(circuit.Y6, circuit.Y6), api.Mul(circuit.Y7, circuit.Y7), api.Mul(circuit.Y8, circuit.Y8), api.Mul(circuit.Y9, circuit.Y9), api.Mul(circuit.Y10, circuit.Y10))
	top := api.Mul(circuit.M1, circuit.M1, circuit.M2, circuit.M2, //m1^2 * m2^2
		api.Sub(api.Add(circuit.M1, circuit.M2), 2), // m1 + m2 - 2
		// m2^2 · (sum(X))^2 + m1^2 * (sum(Y))^2 - 2 * m1 * m2 * sum(X) * sum(Y)
		api.Sub(api.Add(api.Mul(circuit.M2, circuit.M2, sumX, sumX), api.Mul(circuit.M1, circuit.M1, sumY, sumY)), api.Mul(2, circuit.M1, circuit.M2, sumX, sumY)))
	bottom := api.Mul(api.Add(api.Mul(circuit.M1, circuit.M2, circuit.M2), api.Mul(circuit.M1, circuit.M1, circuit.M2)), // (m1m2^2 + m1^2m2)
		// m1 * m2^2 * sum(X^2) -m2^2 * (sum(X))^2 + m1^2 * m2 * sum(Y^2) - m1^2 * (sum(Y))^2
		api.Sub(api.Add(api.Sub(api.Mul(circuit.M1, circuit.M2, circuit.M2, SquareSumX), api.Mul(circuit.M2, circuit.M2, sumX, sumX)), api.Mul(circuit.M1, circuit.M1, circuit.M2, SquareSumY)), api.Mul(circuit.M1, circuit.M1, sumY, sumY)))
	api.AssertIsLessOrEqual(api.Sub(top, api.Mul(bottom, circuit.T)), bottom)
	return nil
}
