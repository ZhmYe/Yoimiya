package Circuit4MatrixMultiplication

import (
	"Yoimiya/frontend"
)

type MatrixMultiplicationCircuit struct {
	A [256][256]frontend.Variable `gnark:",public"`
	B [256][256]frontend.Variable `gnark:",public"`
	C [256][256]frontend.Variable `gnark:"C"`
	//C frontend.Variable `gnark:"C"`
}

func (c *MatrixMultiplicationCircuit) Define(api frontend.API) error {
	//sum := frontend.Variable(0)
	var tmp [256][256]frontend.Variable
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			tmp[i][j] = frontend.Variable(0)
			for k := 0; k < 256; k++ {
				tmp[i][j] = api.Add(tmp[i][j], api.Mul(c.A[i][k], c.B[k][j]))
			}
		}
	}
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			api.AssertIsEqual(c.C[i][j], tmp[i][j])
		}
	}
	return nil
}
