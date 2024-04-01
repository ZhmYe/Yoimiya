package Circuit4Loops

import (
	"S-gnark/backend/groth16"
	"S-gnark/frontend"
	"S-gnark/frontend/cs/r1cs"
	"S-gnark/frontend/split"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
	"time"
)

// }
func TestLoop(t *testing.T) {
	startTime := time.Now()
	var circuit LoopCircuit
	assignment := LoopCircuit{V: 20365011074, X1: 0, X2: 1}
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
}
