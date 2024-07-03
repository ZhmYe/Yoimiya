package Circuit4MatrixMultiplication

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

func TestMatrixMultiplication(t *testing.T) {
	startTime := time.Now()
	var circuit MatrixMultiplicationCircuit
	assignmentGenerator := func() frontend.Circuit {
		var A [150][150]frontend.Variable
		for i := 0; i < 150; i++ {
			//A = append(A, make([]frontend.Variable, 150))
			for j := 0; j < 150; j++ {
				A[i][j] = frontend.Variable(0)
			}
		}
		return &MatrixMultiplicationCircuit{A: A, B: A, C: A}
	}
	assignment := assignmentGenerator()
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		panic(err)
	}
	fmt.Println("Compile Time:", time.Since(startTime))
	proofs, err := split.Split(ccs, assignment, 2)
	if err != nil {
		panic("error")
	}
	for i, packedProof := range proofs {
		proof := packedProof.GetProof()
		verifyKey := packedProof.GetVerifyingKey()
		publicWitness := packedProof.GetPublicWitness()
		err := groth16.Verify(proof, verifyKey, publicWitness)
		if err != nil {
			panic(err)
		}
		fmt.Println("Proof ", i, " Verify Pass...")
	}
	fmt.Println(Record.GlobalRecord)
}
