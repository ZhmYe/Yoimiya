package Circuit4Conv

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

func TestConvolution(t *testing.T) {
	startTime := time.Now()
	var circuit ConvolutionalCircuit
	assignmentGenerator := func() frontend.Circuit {
		var A [256][256]frontend.Variable
		for i := 0; i < 256; i++ {
			//A = append(A, make([]frontend.Variable, 100))
			for j := 0; j < 256; j++ {
				A[i][j] = frontend.Variable(0)
			}
		}
		var W [3][3]frontend.Variable
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				W[i][j] = frontend.Variable(0)
			}
		}
		var B [254][254]frontend.Variable
		for i := 0; i < 254; i++ {
			for j := 0; j < 254; j++ {
				B[i][j] = frontend.Variable(0)
			}
		}
		return &ConvolutionalCircuit{A: A, W: W, B: B}
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
