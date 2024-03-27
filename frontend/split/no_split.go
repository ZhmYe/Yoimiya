package split

import (
	"S-gnark/backend/groth16"
	"S-gnark/constraint"
	"S-gnark/frontend"
	"github.com/consensys/gnark-crypto/ecc"
)

// SimpleProve 不切分电路
func SimpleProve(cs constraint.ConstraintSystem, assignment frontend.Circuit) ([]PackedProof, error) {
	pk, vk := frontend.SetUpSplit(cs)
	extra := make([]constraint.ExtraValue, 0)
	fullWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField())
	publicWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField(), frontend.PublicOnly())
	proof, _ := groth16.Prove(cs, pk.(groth16.ProvingKey), fullWitness)
	return []PackedProof{NewPackedProof(proof, vk, publicWitness)}, nil
}
