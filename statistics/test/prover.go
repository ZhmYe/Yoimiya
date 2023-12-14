package test

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
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
