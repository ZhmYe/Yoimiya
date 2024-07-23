package plugin

import (
	"Yoimiya/backend/groth16"
	groth16_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs "Yoimiya/constraint/bn254"
)

// Prover 一个prover为某一种特定的ccs(pk一致)生成proof
// 接收commitment和solution，以及nbPublic、nbPrivate
type Prover struct {
	pk *groth16_bn254.ProvingKey
}

func NewProver(pk groth16.ProvingKey) Prover {
	return Prover{pk: pk.(*groth16_bn254.ProvingKey)}
}
func (p *Prover) Prove(solution cs.R1CSSolution, commitmentInfo constraint.Groth16Commitments, nbPublic int, nbPrivate int) (groth16.Proof, error) {
	return groth16_bn254.GenerateZKP(commitmentInfo, solution, p.pk, nbPublic, nbPrivate)
}
func (p *Prover) Solve(r1cs *cs.R1CS, fullWitness witness.Witness) (constraint.Groth16Commitments, *cs.R1CSSolution, int, int) {
	commitmentInfo, solution, nbPublic, nbPrivate, err := groth16_bn254.Solve(r1cs, fullWitness, p.pk)
	if err != nil {
		panic(err)
	}
	return commitmentInfo, solution, nbPublic, nbPrivate
}
func (p *Prover) SolveAndProve(r1cs *cs.R1CS, fullWitness witness.Witness) (groth16.Proof, error) {
	commitmentInfo, solution, nbPublic, nbPrivate := p.Solve(r1cs, fullWitness)
	return p.Prove(*solution, commitmentInfo, nbPublic, nbPrivate)
	//return groth16_bn254.Prove(r1cs, p.pk, fullWitness)
}
