package Circuit4VerifyCircuit

import (
	"Yoimiya/backend/groth16"
	"fmt"
	"testing"
	"time"
)

func Test4VerifyCircuit(t *testing.T) {
	innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray := GetVerifyCircuitParam()

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
	for i, packedProof := range proofs {
		proof := packedProof.GetProof()
		verifyKey := packedProof.GetVerifyingKey()
		publicWitness := packedProof.GetPublicWitness()
		err := groth16.Verify(proof, verifyKey, publicWitness)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Proof ", i, " Verify Success...")
		}
	}
}
