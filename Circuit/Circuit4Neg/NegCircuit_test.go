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
		return &NegCircuit{X1: 0, X2: 0, V1: 0, V2: 0}
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
}
