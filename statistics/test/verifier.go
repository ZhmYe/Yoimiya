package test

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"fmt"
)

type Verifier struct {
	vk            groth16.VerifyingKey
	publicWitness witness.Witness
}

func (v *Verifier) verify(proof groth16.Proof) bool {
	err := groth16.Verify(proof, v.vk, v.publicWitness)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
