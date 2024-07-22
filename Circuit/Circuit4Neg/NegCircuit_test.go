package Circuit4Neg

import (
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
	"time"
)

func TestNeg(t *testing.T) {
	startTime := time.Now()
	var circuit NegCircuit
	assignmentGenerator := func() frontend.Circuit {
		return &NegCircuit{X: 1}
	}
	Config.Config.CancelSplit()
	assignment := assignmentGenerator()
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	fmt.Println(ccs.GetNbConstraints())
	if err != nil {
		panic(err)
	}
	fmt.Println("Compile Time:", time.Since(startTime))
	pk, _, err := groth16.Setup(ccs)
	fullWitness, _ := frontend.GenerateWitness(assignment, make([]constraint.ExtraValue, 0), ecc.BN254.ScalarField())

	_, err = groth16.Prove(ccs, pk.(groth16.ProvingKey), fullWitness)
	if err != nil {
		panic(err)
	}
	//proofs, err := split.Split(ccs, assignment, 1)
	//if err != nil {
	//	panic("error")
	//}
	//for i, packedProof := range proofs {
	//	proof := packedProof.GetProof()
	//	verifyKey := packedProof.GetVerifyingKey()
	//	publicWitness := packedProof.GetPublicWitness()
	//	err := groth16.Verify(proof, verifyKey, publicWitness)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println("Proof ", i, " Verify Pass...")
	//}
	//fmt.Println(Record.GlobalRecord)
}
