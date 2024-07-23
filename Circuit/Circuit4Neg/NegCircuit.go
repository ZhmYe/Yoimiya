package Circuit4Neg

import (
	"Yoimiya/Config"
	"Yoimiya/frontend"
)

type NegCircuit struct {
	X1 frontend.Variable `gnark:",public"` // a_1,a_2
	X2 frontend.Variable `gnark:",public"`
	V1 frontend.Variable `gnark:"v1"` // an-1
	V2 frontend.Variable `gnark:"v2"` // an
}

func (c *NegCircuit) Define(api frontend.API) error {
	for i := 0; i < Config.Config.NbLoop; i++ {
		c.X1 = api.Add(api.Mul(c.X1, c.X1), api.Mul(c.X2, c.X2))
		c.X2 = api.Add(api.Mul(c.X1, c.X1), api.Mul(c.X2, c.X2))
	}
	api.AssertIsEqual(c.V1, c.X1)
	api.AssertIsEqual(c.V2, c.X2)
	return nil
}
