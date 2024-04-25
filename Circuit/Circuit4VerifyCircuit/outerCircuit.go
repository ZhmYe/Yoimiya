package Circuit4VerifyCircuit

import (
	"Yoimiya/frontend"
	"Yoimiya/std/algebra"
	"Yoimiya/std/math/emulated"
	"fmt"
)

const LENGTH int = 1

type OuterCircuit[FR emulated.FieldParams, G1El algebra.G1ElementT, G2El algebra.G2ElementT, GtEl algebra.GtElementT] struct {
	Proof        [LENGTH]Proof[G1El, G2El]
	VerifyingKey [LENGTH]VerifyingKey[G1El, G2El, GtEl]
	InnerWitness [LENGTH]Witness[FR]
}

func (c *OuterCircuit[FR, G1El, G2El, GtEl]) Define(api frontend.API) error {
	curve, err := algebra.GetCurve[FR, G1El](api)
	if err != nil {
		return fmt.Errorf("new curve: %w", err)
	}
	pairing, err := algebra.GetPairing[G1El, G2El, GtEl](api)
	if err != nil {
		return fmt.Errorf("get pairing: %w", err)
	}
	verifier := NewVerifier(curve, pairing)
	for i := range c.Proof {
		err = verifier.AssertProof(c.VerifyingKey[i], c.Proof[i], c.InnerWitness[i])
	}
	return err
}
