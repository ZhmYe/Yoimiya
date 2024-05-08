package Circuit4Multiplication

import (
	"Yoimiya/Record"
	"Yoimiya/backend/groth16"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"Yoimiya/frontend/split"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
	"time"
)

func TestLoopMultiplication(t *testing.T) {
	startTime := time.Now()
	var circuit MultiplicationCircuit
	assignment := MultiplicationCircuit{X: 1, Y: 1}
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		panic(err)
	}
	fmt.Println("Compile Time:", time.Since(startTime))
	proofs, err := split.Split(ccs, &assignment, split.NewParam(true, false, -1, false))
	if err != nil {
		panic("error")
	}
	for _, packedProof := range proofs {
		proof := packedProof.GetProof()
		verifyKey := packedProof.GetVerifyingKey()
		publicWitness := packedProof.GetPublicWitness()
		err := groth16.Verify(proof, verifyKey, publicWitness)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(Record.GlobalRecord)
}
