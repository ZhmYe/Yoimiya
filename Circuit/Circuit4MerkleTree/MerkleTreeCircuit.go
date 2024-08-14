package Circuit4MerkleTree

import (
	"Yoimiya/frontend"
	"Yoimiya/std/hash/mimc"
)

// MerkleTreeCircuit 默克尔树
type MerkleTreeCircuit struct {
	X [16]frontend.Variable // 这里方便期间将下面的叶子节点定为2^b
	R frontend.Variable     `gnark:",public"`
	//X1 frontend.Variable `gnark:",public"` // a_1,a_2
	//X2 frontend.Variable `gnark:",public"`
	//V1 frontend.Variable `gnark:"v1"` // an-1
	//V2 frontend.Variable `gnark:"v2"` // an
}

func (c *MerkleTreeCircuit) Define(api frontend.API) error {
	//sum := frontend.Variable(0)
	currentQueue := make([]frontend.Variable, 0)
	nextQueue := make([]frontend.Variable, 0)
	for _, x := range c.X {
		currentQueue = append(currentQueue, x)
	}
	for len(currentQueue) != 1 {
		for i := 0; i < len(currentQueue)/2; i++ {
			first := 2 * i
			second := 2*i + 1
			//if second >= len(currentQueue) {
			//	second =
			//}
			mimc, _ := mimc.NewMiMC(api)

			// specify constraints
			// mimc(preImage) == hash
			mimc.Write(api.Mul(c.X[first], c.X[second]))
			//api.AssertIsEqual(circuit.Hash, mimc.Sum())
			nextQueue = append(nextQueue, mimc.Sum())

		}
		currentQueue = nextQueue
	}
	api.AssertIsEqual(currentQueue[0], c.R)
	return nil
}
