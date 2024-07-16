package Circuir4Mimc

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

func TestPreimage(t *testing.T) {
	startTime := time.Now()
	var circuit MimcCircuit
	assignment := MimcCircuit{
		PreImage: "16130099170765464552823636852555369511329944820189892919423002775646948828469",
		Hash:     "12886436712380113721405259596386800092738845035233065858332878701083870690753",
	}
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		panic(err)
	}
	fmt.Println("Compile Time:", time.Since(startTime))
	proofs, err := split.Split(ccs, &assignment, 2)
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
