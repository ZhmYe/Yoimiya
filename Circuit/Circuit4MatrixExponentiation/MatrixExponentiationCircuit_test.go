package Circuit4MatrixExponentiation

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

func TestMatrixExponentiation(t *testing.T) {
	startTime := time.Now()
	var circuit MatrixExponentiationCircuit
	assignmentGenerator := func() frontend.Circuit {
		var A [2][2]frontend.Variable
		for i := 0; i < 2; i++ {
			//A = append(A, make([]frontend.Variable, 300))
			for j := 0; j < 2; j++ {
				A[i][j] = frontend.Variable(0)
			}
		}
		return &MatrixExponentiationCircuit{X: A, Y: A}
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
