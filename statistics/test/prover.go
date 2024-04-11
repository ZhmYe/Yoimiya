package test

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
)

type Prover struct {
	ccs     constraint.ConstraintSystem
	pk      groth16.ProvingKey
	witness witness.Witness
	proof   groth16.Proof
}

func (p *Prover) prove() {
	p.proof, _ = groth16.Prove(p.ccs, p.pk, p.witness)
}
func (p *Prover) getProof() groth16.Proof {
	return p.proof
}
