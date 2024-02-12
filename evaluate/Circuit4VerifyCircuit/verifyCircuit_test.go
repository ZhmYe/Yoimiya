package Circuit4VerifyCircuit

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
)

func Test4VerifyCircuit(t *testing.T) {

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

	// outer proof
	//outerCcs, outerPK, outerVK, full, public := getCircuitVkWitnessPublic(assert, innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray)
	middleCcs, middlePK, middleVK, middleFull, middlePublic := GetCircuitVkWitnessPublic(innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray)

	middleProof, _ := groth16.Prove(middleCcs, middlePK.(groth16.ProvingKey), middleFull)

	err := groth16.Verify(middleProof, middleVK, middlePublic)
	if err != nil {
		return
	}
}
