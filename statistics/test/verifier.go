package test

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
)

type Verifier struct {
	vk            groth16.VerifyingKey
	publicWitness witness.Witness
}

func (v *Verifier) verify(proof groth16.Proof) bool {
	err := groth16.Verify(proof, v.vk, v.publicWitness)
	if err != nil {
		return false
	}
	return true
}
