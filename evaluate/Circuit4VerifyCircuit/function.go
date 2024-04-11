package Circuit4VerifyCircuit

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	"github.com/consensys/gnark-crypto/ecc"
)

func GetVerifyCircuitParam() ([LENGTH]constraint.ConstraintSystem,
	[LENGTH]groth16.VerifyingKey,
	[LENGTH]witness.Witness,
	[LENGTH]groth16.Proof) {
	var innerCcsArray [LENGTH]constraint.ConstraintSystem
	var innerVKArray [LENGTH]groth16.VerifyingKey
	var innerWitnessArray [LENGTH]witness.Witness
	var innerProofArray [LENGTH]groth16.Proof

	for i := 0; i < LENGTH; i++ {
		innerCcs, innerVK, innerWitness, innerProof := GetInner(ecc.BN254.ScalarField())
		innerCcsArray[i] = innerCcs
		innerVKArray[i] = innerVK
		innerWitnessArray[i] = innerWitness
		innerProofArray[i] = innerProof
	}
	return innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray
}
