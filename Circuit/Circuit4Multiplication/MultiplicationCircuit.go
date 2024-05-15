package Circuit4Multiplication

import (
	"Yoimiya/Config"
	"Yoimiya/frontend"
)

type MultiplicationCircuit struct {
	//x frontend.Variable
	//X1, X2 frontend.Variable `gnark:"public"` // a_1,a_2
	////N      frontend.Variable `gnark:"public"` // 循环的次数
	//V1 frontend.Variable `gnark:"v1"` // an-1
	//V2 frontend.Variable `gnark:"v2"`
	X frontend.Variable `gnark:",public"`
	Y frontend.Variable `gnark:"y"`
}

func (c *MultiplicationCircuit) Define(api frontend.API) error {
	for i := 0; i < Config.Config.NbLoop; i++ {
		c.X = api.Mul(c.X, c.X)
	}
	api.AssertIsEqual(c.X, c.Y)
	return nil
}
