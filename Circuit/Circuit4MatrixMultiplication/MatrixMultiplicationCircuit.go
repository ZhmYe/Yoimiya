package Circuit4MatrixMultiplication

import (
	"Yoimiya/frontend"
)

type MatrixMultiplicationCircuit struct {
	A [150][150]frontend.Variable `gnark:",public"`
	B [150][150]frontend.Variable `gnark:",public"`
	C [150][150]frontend.Variable `gnark:"C"`
	//C frontend.Variable `gnark:"C"`
}

func (c *MatrixMultiplicationCircuit) Define(api frontend.API) error {
	//sum := frontend.Variable(0)
	var tmp [150][150]frontend.Variable
	for i := 0; i < 150; i++ {
		for j := 0; j < 150; j++ {
			tmp[i][j] = frontend.Variable(0)
			for k := 0; k < 150; k++ {
				tmp[i][j] = api.Add(tmp[i][j], api.Mul(c.A[i][k], c.B[k][j]))
			}
		}
	}
	for i := 0; i < 150; i++ {
		for j := 0; j < 150; j++ {
			api.AssertIsEqual(c.C[i][j], tmp[i][j])
		}
	}
	return nil
}
