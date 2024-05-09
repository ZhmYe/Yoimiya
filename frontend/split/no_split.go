package split

import (
	"Yoimiya/Record"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"github.com/consensys/gnark-crypto/ecc"
	"time"
)

// SimpleProve 不切分电路
func SimpleProve(cs constraint.ConstraintSystem, assignment frontend.Circuit) ([]PackedProof, error) {
	startTime := time.Now()
	pk, vk := frontend.SetUpSplit(cs)
	Record.GlobalRecord.SetSetUpTime(time.Since(startTime))
	startTime = time.Now()
	//fmt.Println("Set Up Time: ", time.Since(startTime))
	extra := make([]constraint.ExtraValue, 0)
	fullWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField())
	publicWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField(), frontend.PublicOnly())
	//proof, _ := groth16.Prove(cs, pk.(groth16.ProvingKey), fullWitness)
	_, err := groth16.Prove(cs, pk.(groth16.ProvingKey), fullWitness)
	if err != nil {
		return nil, err
	}
	Record.GlobalRecord.SetSolveTime(time.Since(startTime))
	return []PackedProof{NewPackedProof(nil, vk, publicWitness)}, nil
}
