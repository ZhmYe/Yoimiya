package Circuit4Loops

import "Yoimiya/frontend"

type LoopCircuit struct {
	//x frontend.Variable
	X1, X2 frontend.Variable
	V      frontend.Variable `gnark:"v"`
}

func (c *LoopCircuit) Define(api frontend.API) error {
	C := frontend.Variable(0)
	L, R := c.X1, c.X2
	for i := 0; i < 49; i++ {
		C = api.Add(L, R)
		L = R
		R = C
	}
	api.AssertIsEqual(c.V, api.Add(L, R))
	return nil
}
