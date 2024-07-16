package Split

import (
	"Yoimiya/backend/groth16"
	groth16_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/backend/witness"
	fr_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type Verifier struct {
	vk            *groth16_bn254.VerifyingKey
	publicWitness witness.Witness
}

func NewVerifier(vk groth16.VerifyingKey, w witness.Witness) Verifier {
	return Verifier{vk: vk.(*groth16_bn254.VerifyingKey), publicWitness: w}
}
func (v *Verifier) Verify(proof groth16.Proof) (bool, error) {
	err := groth16_bn254.Verify(proof.(*groth16_bn254.Proof), v.vk, v.publicWitness.Vector().(fr_bn254.Vector))
	if err != nil {
		return false, err
	}
	return true, nil
}
