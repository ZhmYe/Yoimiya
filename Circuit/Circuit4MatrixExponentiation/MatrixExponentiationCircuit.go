package Circuit4MatrixExponentiation

import (
	"Yoimiya/frontend"
)

// MatrixExponentiationCircuit 矩阵求幂
type MatrixExponentiationCircuit struct {
	X [2][2]frontend.Variable `gnark:",public"`
	Y [2][2]frontend.Variable `gnark:"Y"`
	//X1    frontend.Variable   `gnark:",public"` // a_1,a_2
	//X2    frontend.Variable   `gnark:",public"`
	//V1    frontend.Variable   `gnark:"v1"` // an-1
	//V2    frontend.Variable   `gnark:"v2"` // an
}

func (c *MatrixExponentiationCircuit) Define(api frontend.API) error {
	var result [2][2]frontend.Variable
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			result[i][j] = frontend.Variable(0)
		}
	}
	for x := 0; x < 10000; x++ {
		var tmp [2][2]frontend.Variable
		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				tmp[i][j] = frontend.Variable(0)
				for k := 0; k < 2; k++ {
					tmp[i][j] = api.Add(tmp[i][j], api.Mul(result[i][k], c.X[k][j]))
				}
			}
		}
		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				result[i][j] = api.Mul(1, tmp[i][j])
			}
		}
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			api.AssertIsEqual(c.Y[i][j], result[i][j])
		}
	}
	return nil
	//for i := 0; i < Config.Config.NbLoop; i++ {
	//	c.X1 = api.Add(api.Mul(c.X1, c.X1), api.Mul(c.X2, c.X2))
	//	c.X2 = api.Add(api.Mul(c.X1, c.X1), api.Mul(c.X2, c.X2))
	//}
	//api.AssertIsEqual(c.V1, c.X1)
	//api.AssertIsEqual(c.V2, c.X2)
	//return nil
}
