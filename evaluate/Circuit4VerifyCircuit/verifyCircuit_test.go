package Circuit4VerifyCircuit

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
	"time"
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
	//middleCcs, middlePK, middleVK, middleFull, middlePublic := GetCircuitVkWitnessPublic(innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray)
	//
	//middleProof, _ := groth16.Prove(middleCcs, middlePK.(groth16.ProvingKey), middleFull)
	//
	//err := groth16.Verify(middleProof, middleVK, middlePublic)
	//if err != nil {
	//	return
	//}
	startTime := time.Now()
	proofs := GetPackProofInSplit(innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray)
	fmt.Println(len(proofs))
	fmt.Println("Split Circuit Time:", time.Since(startTime))
	for _, packedProof := range proofs {
		proof := packedProof.GetProof()
		verifyKey := packedProof.GetVerifyingKey()
		publicWitness := packedProof.GetPublicWitness()
		err := groth16.Verify(proof, verifyKey, publicWitness)
		if err != nil {
			panic(err)
		}
	}
}
